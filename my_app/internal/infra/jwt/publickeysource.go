package jwt

import (
	"context"
	"crypto/rsa"
)

type PublicKeySource interface {
	FindPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error)
	ListAllPublicKeys() []*rsa.PublicKey
}
