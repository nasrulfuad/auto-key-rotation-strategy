package signingkey

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type Signer struct {
	signingKeyReader port.SigningKeyReader
	tokenIssuer      port.TokenIssuer
	clock            port.Clock
	uuidGenerator    port.UUIDGenerator

	issuer   string
	tokenTTL time.Duration
}

func NewSigner(
	signingKeyReader port.SigningKeyReader,
	tokenIssuer port.TokenIssuer,
	clock port.Clock,
	uuidGenerator port.UUIDGenerator,
	issuer string,
	tokenTTL time.Duration,
) *Signer {
	return &Signer{
		signingKeyReader: signingKeyReader,
		tokenIssuer:      tokenIssuer,
		clock:            clock,
		uuidGenerator:    uuidGenerator,
		issuer:           issuer,
		tokenTTL:         tokenTTL,
	}
}

func (s *Signer) Sign(ctx context.Context, request domain.TokenRequest) (domain.SignedToken, error) {
	now := s.clock.Now()
	key, err := s.signingKeyReader.FindActive(ctx, now)
	if err != nil {
		return domain.SignedToken{}, err
	}
	issuedAt := now.Truncate(time.Second)
	claims := domain.TokenClaims{
		ID:        s.uuidGenerator.NewUUID().String(),
		Issuer:    s.issuer,
		Subject:   request.Subject,
		Provider:  request.Provider,
		Audience:  []string{string(request.Audience)},
		IssuedAt:  issuedAt,
		ExpiresAt: issuedAt.Add(s.tokenTTL),
	}
	value, err := s.tokenIssuer.Issue(ctx, claims, key)
	if err != nil {
		return domain.SignedToken{}, domain.Internal(err)
	}
	return domain.SignedToken{Value: value, KeyID: key.KeyID, Claims: claims, ExpiresAt: claims.ExpiresAt}, nil
}
