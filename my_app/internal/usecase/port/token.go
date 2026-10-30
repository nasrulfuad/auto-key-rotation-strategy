package port

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type TokenIssuer interface {
	Issue(ctx context.Context, claims domain.TokenClaims, key domain.SigningKey) (string, error)
}

type TokenVerifier interface {
	Verify(ctx context.Context, raw string, expectation domain.TokenExpectation, now time.Time) (domain.VerifiedToken, error)
}

type PeerKeySet interface {
	ListAll(ctx context.Context) []domain.PeerSigningKey
	Refresh(ctx context.Context) error
	Status() domain.PeerKeySetStatus
}
