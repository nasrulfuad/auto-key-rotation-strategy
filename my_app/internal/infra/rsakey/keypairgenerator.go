package rsakey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

// KeyPairGenerator creates new signing key pairs.
type KeyPairGenerator interface {
	// Generate creates an RSA key pair for algorithm, with its RFC 7638 thumbprint as key ID and the private key PKCS#8 encoded.
	// It fails for an unsupported algorithm.
	Generate(ctx context.Context, algorithm domain.SigningAlgorithm) (domain.KeyPair, error)
}
