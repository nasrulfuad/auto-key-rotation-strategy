package signingkey

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwt"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/systemclock"
)

type Signer struct {
	signingKeyReader repository.SigningKeyReader
	tokenIssuer      jwt.TokenIssuer
	clock            systemclock.Clock
	uuidGenerator    systemclock.UUIDGenerator

	issuer   string
	tokenTTL time.Duration
}

func NewSigner(
	signingKeyReader repository.SigningKeyReader,
	tokenIssuer jwt.TokenIssuer,
	issuer string,
	tokenTTL time.Duration,
	opts ...Option,
) *Signer {
	o := newOptions(opts)
	return &Signer{
		signingKeyReader: signingKeyReader,
		tokenIssuer:      tokenIssuer,
		clock:            o.clock,
		uuidGenerator:    o.uuidGenerator,
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
