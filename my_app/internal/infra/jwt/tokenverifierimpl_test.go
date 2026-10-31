package jwt_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwks"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwt"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/rsakey"
)

var now = time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

type fixedClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fixedClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fixedClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

type peerKey struct {
	keyID      string
	privateKey *rsa.PrivateKey
}

type peer struct {
	mu      sync.Mutex
	keys    []peerKey
	fetches atomic.Int32
	server  *httptest.Server
}

func newPeer(t *testing.T) *peer {
	p := &peer{}
	p.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		p.fetches.Add(1)
		p.mu.Lock()
		defer p.mu.Unlock()
		type key struct {
			KeyType   string `json:"kty"`
			Use       string `json:"use"`
			KeyID     string `json:"kid"`
			Algorithm string `json:"alg"`
			Modulus   string `json:"n"`
			Exponent  string `json:"e"`
		}
		document := struct {
			Keys []key `json:"keys"`
		}{}
		for _, k := range p.keys {
			public := rsakey.ToDomainPublicKey(&k.privateKey.PublicKey)
			document.Keys = append(document.Keys, key{"RSA", "sig", k.keyID, "RS256",
				rsakey.EncodeBase64URL(public.Modulus), rsakey.EncodeBase64URL(public.Exponent)})
		}
		_ = json.NewEncoder(w).Encode(document)
	}))
	t.Cleanup(p.server.Close)
	return p
}

func (p *peer) addKey(t *testing.T, keyID string) peerKey {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	k := peerKey{keyID: keyID, privateKey: privateKey}
	p.mu.Lock()
	p.keys = append(p.keys, k)
	p.mu.Unlock()
	return k
}

func (p *peer) removeKey(keyID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	for i, k := range p.keys {
		if k.keyID == keyID {
			p.keys = append(p.keys[:i], p.keys[i+1:]...)
			return
		}
	}
}

func validClaims() domain.TokenClaims {
	return domain.TokenClaims{
		ID:        "jti",
		Issuer:    "qa.peer.local",
		Provider:  "my_app",
		Audience:  []string{"get.entitlement"},
		IssuedAt:  now,
		ExpiresAt: now.Add(5 * time.Minute),
	}
}

var expectation = domain.TokenExpectation{Issuer: "qa.peer.local", Audience: "get.entitlement", Leeway: 30 * time.Second}

func sign(t *testing.T, claims domain.TokenClaims, key peerKey) string {
	token, err := jwt.Sign(claims, key.keyID, domain.SigningAlgorithmRS256, key.privateKey)
	require.NoError(t, err)
	return token
}

func newVerifier(p *peer, clock *fixedClock, refreshOnUnknownKeyID bool) (*jwt.Verifier, *jwks.RemoteKeySet) {
	keySet := jwks.NewRemoteKeySet(p.server.URL, jwks.WithRefreshOnUnknownKeyID(refreshOnUnknownKeyID), jwks.WithRefreshMinimumInterval(10*time.Second), jwks.WithClock(clock))
	return jwt.NewVerifier(keySet, domain.SigningAlgorithmRS256), keySet
}

func TestVerifierAcceptsTokenAndReportsKeyID(t *testing.T) {
	p := newPeer(t)
	key := p.addKey(t, "k1")
	verifier, _ := newVerifier(p, &fixedClock{now: now}, true)

	verified, err := verifier.Verify(context.Background(), sign(t, validClaims(), key), expectation, now)

	require.NoError(t, err)
	assert.Equal(t, "k1", verified.KeyID)
	assert.Equal(t, "my_app", verified.Claims.Provider)
}

func TestVerifierRefreshesOnceForUnknownKeyID(t *testing.T) {
	p := newPeer(t)
	first := p.addKey(t, "k1")
	clock := &fixedClock{now: now}
	verifier, keySet := newVerifier(p, clock, true)
	require.NoError(t, keySet.Refresh(context.Background()))
	second := p.addKey(t, "k2")
	clock.Advance(11 * time.Second)

	_, err := verifier.Verify(context.Background(), sign(t, validClaims(), second), expectation, now)
	require.NoError(t, err)
	_, err = verifier.Verify(context.Background(), sign(t, validClaims(), first), expectation, now)
	require.NoError(t, err)

	assert.Equal(t, int32(2), p.fetches.Load())
}

func TestVerifierRespectsRefreshMinimumInterval(t *testing.T) {
	p := newPeer(t)
	p.addKey(t, "k1")
	clock := &fixedClock{now: now}
	verifier, keySet := newVerifier(p, clock, true)
	require.NoError(t, keySet.Refresh(context.Background()))
	second := p.addKey(t, "k2")

	_, err := verifier.Verify(context.Background(), sign(t, validClaims(), second), expectation, now)

	assert.ErrorIs(t, err, domain.ErrTokenSigningKeyNotFound)
	assert.Equal(t, int32(1), p.fetches.Load())
}

func TestVerifierWithoutRefreshRejectsNewKeyUntilPolled(t *testing.T) {
	p := newPeer(t)
	p.addKey(t, "k1")
	clock := &fixedClock{now: now}
	verifier, keySet := newVerifier(p, clock, false)
	require.NoError(t, keySet.Refresh(context.Background()))
	second := p.addKey(t, "k2")
	token := sign(t, validClaims(), second)

	_, err := verifier.Verify(context.Background(), token, expectation, now)
	assert.ErrorIs(t, err, domain.ErrTokenSigningKeyNotFound)

	require.NoError(t, keySet.Refresh(context.Background()))
	_, err = verifier.Verify(context.Background(), token, expectation, now)
	assert.NoError(t, err)
}

func TestVerifierRejectsTokenOfRemovedKey(t *testing.T) {
	p := newPeer(t)
	old := p.addKey(t, "k1")
	p.addKey(t, "k2")
	clock := &fixedClock{now: now}
	verifier, keySet := newVerifier(p, clock, true)
	require.NoError(t, keySet.Refresh(context.Background()))
	token := sign(t, validClaims(), old)
	p.removeKey("k1")
	require.NoError(t, keySet.Refresh(context.Background()))

	_, err := verifier.Verify(context.Background(), token, expectation, now)

	assert.ErrorIs(t, err, domain.ErrTokenSigningKeyNotFound)
}

func TestVerifierFallsBackToEveryKeyWhenKidIsMissing(t *testing.T) {
	p := newPeer(t)
	key := p.addKey(t, "k1")
	verifier, keySet := newVerifier(p, &fixedClock{now: now}, false)
	require.NoError(t, keySet.Refresh(context.Background()))
	token := gojwt.NewWithClaims(gojwt.SigningMethodRS256, gojwt.MapClaims{
		"iss": "qa.peer.local", "aud": []string{"get.entitlement"}, "iat": now.Unix(), "exp": now.Add(time.Minute).Unix(),
	})
	raw, err := token.SignedString(key.privateKey)
	require.NoError(t, err)

	_, err = verifier.Verify(context.Background(), raw, expectation, now)

	assert.NoError(t, err)
}

func TestVerifierRejectsUnsafeAndInvalidTokens(t *testing.T) {
	p := newPeer(t)
	key := p.addKey(t, "k1")
	verifier, _ := newVerifier(p, &fixedClock{now: now}, true)

	none := gojwt.NewWithClaims(gojwt.SigningMethodNone, gojwt.MapClaims{"iss": "qa.peer.local"})
	noneRaw, err := none.SignedString(gojwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)
	hmac := gojwt.NewWithClaims(gojwt.SigningMethodHS256, gojwt.MapClaims{"iss": "qa.peer.local"})
	hmacRaw, err := hmac.SignedString([]byte("secret"))
	require.NoError(t, err)
	wrongIssuer := validClaims()
	wrongIssuer.Issuer = "prod.peer.local"
	wrongAudience := validClaims()
	wrongAudience.Audience = []string{"set.entitlement"}
	expired := validClaims()
	expired.ExpiresAt = now.Add(-time.Minute)
	future := validClaims()
	future.IssuedAt = now.Add(time.Minute)
	future.ExpiresAt = now.Add(2 * time.Minute)

	cases := []struct {
		name string
		raw  string
		want error
	}{
		{"alg none", noneRaw, domain.ErrTokenInvalid},
		{"alg HS256", hmacRaw, domain.ErrTokenInvalid},
		{"garbage", "not.a.token", domain.ErrTokenInvalid},
		{"wrong issuer", sign(t, wrongIssuer, key), domain.ErrTokenIssuerUnknown},
		{"wrong audience", sign(t, wrongAudience, key), domain.ErrTokenClaimInvalid},
		{"expired", sign(t, expired, key), domain.ErrTokenClaimInvalid},
		{"issued in the future", sign(t, future, key), domain.ErrTokenClaimInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := verifier.Verify(context.Background(), tc.raw, expectation, now)
			assert.ErrorIs(t, err, tc.want)
		})
	}
}

func TestRemoteKeySetKeepsCachedKeysWhenRefreshFails(t *testing.T) {
	p := newPeer(t)
	key := p.addKey(t, "k1")
	verifier, keySet := newVerifier(p, &fixedClock{now: now}, true)
	require.NoError(t, keySet.Refresh(context.Background()))
	p.server.Close()

	assert.Error(t, keySet.Refresh(context.Background()))
	_, err := verifier.Verify(context.Background(), sign(t, validClaims(), key), expectation, now)

	assert.NoError(t, err)
	assert.NotEmpty(t, keySet.Status().LastError)
}
