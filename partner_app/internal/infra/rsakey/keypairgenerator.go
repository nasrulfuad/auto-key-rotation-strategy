package rsakey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type KeyPairGenerator interface {
	Generate(ctx context.Context, algorithm domain.SigningAlgorithm) (domain.KeyPair, error)
}
