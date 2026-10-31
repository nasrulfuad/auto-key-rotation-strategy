package jwt

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

// TokenVerifier verifies the JWTs signed by the peer.
type TokenVerifier interface {
	// Verify checks the signature of raw against the peer key set and its claims against expectation at now.
	// A rejected token returns domain.ErrTokenIssuerUnknown, domain.ErrTokenSigningKeyNotFound,
	// a TOKEN_INVALID error or a TOKEN_CLAIM_INVALID error.
	Verify(ctx context.Context, raw string, expectation domain.TokenExpectation, now time.Time) (domain.VerifiedToken, error)
}
