package activationnotification

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type TokenSigner interface {
	Sign(ctx context.Context, request domain.TokenRequest) (domain.SignedToken, error)
}
