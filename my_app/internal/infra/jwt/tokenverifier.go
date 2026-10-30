package jwt

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type TokenVerifier interface {
	Verify(ctx context.Context, raw string, expectation domain.TokenExpectation, now time.Time) (domain.VerifiedToken, error)
}
