package authentication

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

type AuthenticatePeerRequestInput struct {
	RawToken string
	Audience domain.TokenAudience
}

type AuthenticatePeerRequestInteractor struct {
	tokenVerifier port.TokenVerifier
	clock         port.Clock

	peerIssuer string
	leeway     time.Duration
}

func NewAuthenticatePeerRequestInteractor(
	tokenVerifier port.TokenVerifier,
	clock port.Clock,
	peerIssuer string,
	leeway time.Duration,
) *AuthenticatePeerRequestInteractor {
	return &AuthenticatePeerRequestInteractor{tokenVerifier: tokenVerifier, clock: clock, peerIssuer: peerIssuer, leeway: leeway}
}

func (i *AuthenticatePeerRequestInteractor) Execute(ctx context.Context, in AuthenticatePeerRequestInput) (domain.VerifiedToken, error) {
	if in.RawToken == "" {
		return domain.VerifiedToken{}, domain.TokenInvalid("token_missing")
	}
	return i.tokenVerifier.Verify(ctx, in.RawToken, domain.TokenExpectation{
		Issuer:   i.peerIssuer,
		Audience: in.Audience,
		Leeway:   i.leeway,
	}, i.clock.Now())
}
