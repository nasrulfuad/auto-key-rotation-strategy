package jwt

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
)

// TokenIssuer turns claims into a signed JWT.
type TokenIssuer interface {
	// Issue signs claims with the private key of key and returns the compact JWT, with key.KeyID as the kid header.
	Issue(ctx context.Context, claims domain.TokenClaims, key domain.SigningKey) (string, error)
}
