package activationnotification

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
)

// TokenSigner signs tokens with the signing key that is active now.
type TokenSigner interface {
	// Sign issues a token for request, valid for the configured token TTL.
	// It returns domain.ErrSigningKeyNotFound when no key is active.
	Sign(ctx context.Context, request domain.TokenRequest) (domain.SignedToken, error)
}
