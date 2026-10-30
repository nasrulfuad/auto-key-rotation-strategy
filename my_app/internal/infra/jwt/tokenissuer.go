package jwt

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type TokenIssuer interface {
	Issue(ctx context.Context, claims domain.TokenClaims, key domain.SigningKey) (string, error)
}
