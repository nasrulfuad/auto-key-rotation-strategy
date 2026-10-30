package jwks

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/rsakey"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock"
)

const (
	maximumDocumentBytes       = 1 << 20
	failedRefreshRetryInterval = time.Second
)

var ErrKeyNotFound = errors.New("signing key not found in the peer key set")

type Options struct {
	URL                    string
	Client                 *http.Client
	RefreshOnUnknownKeyID  bool
	RefreshMinimumInterval time.Duration
}

type peerKey struct {
	signingKey domain.PeerSigningKey
	publicKey  *rsa.PublicKey
}

type RemoteKeySet struct {
	options Options
	clock   systemclock.Clock
	logger  logging.Logger
	group   singleflight.Group

	mu            sync.RWMutex
	keys          map[string]peerKey
	lastFetchedAt time.Time
	lastAttemptAt time.Time
	lastError     string
}

var _ PeerKeySet = (*RemoteKeySet)(nil)

func NewRemoteKeySet(options Options, clock systemclock.Clock, logger logging.Logger) *RemoteKeySet {
	if options.Client == nil {
		options.Client = &http.Client{Timeout: 5 * time.Second}
	}
	if logger == nil {
		logger = logging.NopLogger{}
	}
	return &RemoteKeySet{options: options, clock: clock, logger: logger, keys: map[string]peerKey{}}
}

func (s *RemoteKeySet) FindPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error) {
	if key, found := s.lookup(keyID); found {
		return key, nil
	}
	if !s.options.RefreshOnUnknownKeyID || !s.refreshAllowed() {
		return nil, ErrKeyNotFound
	}
	s.logger.Info(ctx, "unknown peer kid; refreshing peer key set", "kid", keyID, "url", s.options.URL)
	if err := s.Refresh(ctx); err != nil {
		return nil, err
	}
	if key, found := s.lookup(keyID); found {
		return key, nil
	}
	return nil, ErrKeyNotFound
}

func (s *RemoteKeySet) ListAllPublicKeys() []*rsa.PublicKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	publicKeys := make([]*rsa.PublicKey, 0, len(s.keys))
	for _, key := range s.keys {
		publicKeys = append(publicKeys, key.publicKey)
	}
	return publicKeys
}

func (s *RemoteKeySet) ListAll(_ context.Context) []domain.PeerSigningKey {
	s.mu.RLock()
	defer s.mu.RUnlock()
	signingKeys := make([]domain.PeerSigningKey, 0, len(s.keys))
	for _, key := range s.keys {
		signingKeys = append(signingKeys, key.signingKey)
	}
	slices.SortFunc(signingKeys, func(a, b domain.PeerSigningKey) int {
		if c := a.FetchedAt.Compare(b.FetchedAt); c != 0 {
			return c
		}
		return compareStrings(a.KeyID, b.KeyID)
	})
	return signingKeys
}

func (s *RemoteKeySet) Status() domain.PeerKeySetStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return domain.PeerKeySetStatus{
		URL:           s.options.URL,
		LastFetchedAt: s.lastFetchedAt,
		LastAttemptAt: s.lastAttemptAt,
		LastError:     s.lastError,
	}
}

func (s *RemoteKeySet) Refresh(ctx context.Context) error {
	_, err, _ := s.group.Do("refresh", func() (any, error) {
		return nil, s.refresh(ctx)
	})
	return err
}

func (s *RemoteKeySet) refresh(ctx context.Context) error {
	now := s.clock.Now()
	s.mu.Lock()
	s.lastAttemptAt = now
	s.mu.Unlock()

	fetched, err := s.fetch(ctx)
	if err != nil {
		s.mu.Lock()
		s.lastError = err.Error()
		s.mu.Unlock()
		s.logger.Warn(ctx, "failed to refresh peer key set; keeping the cached keys", "url", s.options.URL, "error", err)
		return err
	}

	s.mu.Lock()
	previous := s.keys
	next := make(map[string]peerKey, len(fetched))
	for keyID, key := range fetched {
		if existing, found := previous[keyID]; found {
			key.signingKey.FetchedAt = existing.signingKey.FetchedAt
		} else {
			key.signingKey.FetchedAt = now
		}
		next[keyID] = key
	}
	s.keys = next
	s.lastFetchedAt = now
	s.lastError = ""
	s.mu.Unlock()

	for keyID := range next {
		if _, found := previous[keyID]; !found {
			s.logger.Info(ctx, "peer signing key discovered in the peer JWKS; tokens signed by it are now accepted", "kid", keyID, "url", s.options.URL, "peer_key_count", len(next))
		}
	}
	for keyID := range previous {
		if _, found := next[keyID]; !found {
			s.logger.Info(ctx, "peer signing key removed from the peer JWKS; tokens signed by it are now rejected", "kid", keyID, "url", s.options.URL, "peer_key_count", len(next))
		}
	}
	return nil
}

type keySetDocument struct {
	Keys []keyDocument `json:"keys"`
}

type keyDocument struct {
	KeyType   string `json:"kty"`
	Use       string `json:"use"`
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
	Exponent  string `json:"e"`
}

func (s *RemoteKeySet) fetch(ctx context.Context) (map[string]peerKey, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, s.options.URL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Accept", "application/json")
	response, err := s.options.Client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("peer key set responded with status %d", response.StatusCode)
	}
	var document keySetDocument
	if err := json.NewDecoder(io.LimitReader(response.Body, maximumDocumentBytes)).Decode(&document); err != nil {
		return nil, fmt.Errorf("failed to decode peer key set: %w", err)
	}
	keys := make(map[string]peerKey, len(document.Keys))
	for _, key := range document.Keys {
		parsed, err := parseKey(key)
		if err != nil {
			s.logger.Warn(ctx, "ignoring invalid peer signing key", "kid", key.KeyID, "error", err)
			continue
		}
		keys[key.KeyID] = parsed
	}
	return keys, nil
}

func parseKey(key keyDocument) (peerKey, error) {
	if key.KeyType != "RSA" {
		return peerKey{}, fmt.Errorf("unsupported key type %q", key.KeyType)
	}
	if key.Use != "" && key.Use != "sig" {
		return peerKey{}, fmt.Errorf("unsupported key use %q", key.Use)
	}
	if key.KeyID == "" {
		return peerKey{}, errors.New("kid is required")
	}
	modulus, err := rsakey.DecodeBase64URL(key.Modulus)
	if err != nil {
		return peerKey{}, fmt.Errorf("invalid modulus: %w", err)
	}
	exponent, err := rsakey.DecodeBase64URL(key.Exponent)
	if err != nil {
		return peerKey{}, fmt.Errorf("invalid exponent: %w", err)
	}
	publicKey := domain.PublicKey{Modulus: modulus, Exponent: exponent}
	rsaPublicKey, err := rsakey.ToRSAPublicKey(publicKey)
	if err != nil {
		return peerKey{}, err
	}
	return peerKey{
		signingKey: domain.PeerSigningKey{KeyID: key.KeyID, Algorithm: domain.SigningAlgorithm(key.Algorithm), PublicKey: publicKey},
		publicKey:  rsaPublicKey,
	}, nil
}

func (s *RemoteKeySet) lookup(keyID string) (*rsa.PublicKey, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key, found := s.keys[keyID]
	return key.publicKey, found
}

func (s *RemoteKeySet) refreshAllowed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lastAttemptAt.IsZero() {
		return true
	}
	interval := s.options.RefreshMinimumInterval
	if s.lastError != "" {
		interval = min(interval, failedRefreshRetryInterval)
	}
	return s.clock.Now().Sub(s.lastAttemptAt) >= interval
}

func compareStrings(a, b string) int {
	switch {
	case a < b:
		return -1
	case a > b:
		return 1
	default:
		return 0
	}
}
