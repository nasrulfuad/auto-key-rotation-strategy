package authentication_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwt/jwtmocks"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/authentication"
)

const peerIssuer = "qa.peer.local"

var verifiedAt = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func TestAuthenticatePeerRequestVerifiesTokenAgainstPeerIssuerAudienceAndLeeway(t *testing.T) {
	verifier := jwtmocks.NewMockTokenVerifier(t)
	clock := systemclockmocks.NewMockClock(t)
	verified := domain.VerifiedToken{KeyID: "peer-kid", Claims: domain.TokenClaims{Issuer: peerIssuer, Provider: "peer"}}
	clock.EXPECT().Now().Return(verifiedAt)
	verifier.EXPECT().Verify(mock.Anything, "raw.jwt.token", domain.TokenExpectation{
		Issuer:   peerIssuer,
		Audience: "activation-notification",
		Leeway:   30 * time.Second,
	}, verifiedAt).Return(verified, nil)

	result, err := authentication.NewAuthenticatePeerRequestInteractor(verifier, peerIssuer, 30*time.Second, authentication.WithClock(clock)).Execute(context.Background(),
		authentication.AuthenticatePeerRequestInput{RawToken: "raw.jwt.token", Audience: "activation-notification"})

	require.NoError(t, err)
	assert.Equal(t, verified, result)
}

func TestAuthenticatePeerRequestRejectsMissingTokenWithoutVerifying(t *testing.T) {
	uc := authentication.NewAuthenticatePeerRequestInteractor(jwtmocks.NewMockTokenVerifier(t), peerIssuer, time.Second, authentication.WithClock(systemclockmocks.NewMockClock(t)))

	_, err := uc.Execute(context.Background(), authentication.AuthenticatePeerRequestInput{Audience: "activation-notification"})

	require.ErrorIs(t, err, domain.TokenInvalid("token_missing"))
}

func TestAuthenticatePeerRequestReturnsVerifierRejection(t *testing.T) {
	verifier := jwtmocks.NewMockTokenVerifier(t)
	clock := systemclockmocks.NewMockClock(t)
	clock.EXPECT().Now().Return(verifiedAt)
	verifier.EXPECT().Verify(mock.Anything, "raw.jwt.token", mock.Anything, verifiedAt).
		Return(domain.VerifiedToken{}, domain.ErrTokenSigningKeyNotFound)

	_, err := authentication.NewAuthenticatePeerRequestInteractor(verifier, peerIssuer, time.Second, authentication.WithClock(clock)).Execute(context.Background(),
		authentication.AuthenticatePeerRequestInput{RawToken: "raw.jwt.token", Audience: "activation-notification"})

	require.ErrorIs(t, err, domain.ErrTokenSigningKeyNotFound)
}
