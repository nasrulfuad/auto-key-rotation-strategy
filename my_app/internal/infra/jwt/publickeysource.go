package jwt

import (
	"context"
	"crypto/rsa"
)

// PublicKeySource provides the peer public keys a Verifier checks signatures with.
type PublicKeySource interface {
	// FindPublicKey returns the public key of keyID, or an error when the key is not known.
	FindPublicKey(ctx context.Context, keyID string) (*rsa.PublicKey, error)

	// ListAllPublicKeys returns every known public key. Verifier tries each of them when the kid is unknown.
	ListAllPublicKeys() []*rsa.PublicKey
}
