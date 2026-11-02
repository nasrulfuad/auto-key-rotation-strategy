package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
)

type SigningKeyCreate struct {
	ID      uuid.UUID
	KeyPair domain.KeyPair
	Window  domain.SigningKeyWindow
}

type SigningKeyRotate struct {
	CurrentKeyID     uuid.UUID
	CurrentKeyWindow domain.SigningKeyWindow
	NextKey          SigningKeyCreate
}

// SigningKeyReader reads the signing keys of this app.
type SigningKeyReader interface {
	// FindLatest returns the key with the latest active_from, or nil and no error when no key exists yet.
	FindLatest(ctx context.Context) (*domain.SigningKey, error)

	// FindActive returns the only key that can sign at now, or domain.ErrSigningKeyNotFound when no key is active.
	FindActive(ctx context.Context, now time.Time) (domain.SigningKey, error)

	// FindPrivateKey decrypts and returns the PKCS#8 private key of keyID, or domain.ErrSigningKeyNotFound when the key does not exist.
	FindPrivateKey(ctx context.Context, keyID string) (domain.Secret, error)

	// ListAllPublished returns the keys published in the JWKS at now (pending, active and inactive), ordered by active_from.
	ListAllPublished(ctx context.Context, now time.Time) ([]domain.SigningKey, error)

	// ListAll returns every stored key, including expired keys not deleted yet, ordered by active_from.
	ListAll(ctx context.Context) ([]domain.SigningKey, error)
}

// SigningKeyWriter stores, rotates and deletes the signing keys of this app.
type SigningKeyWriter interface {
	// Create encrypts the private key of in.KeyPair and stores the key with in.Window.
	Create(ctx context.Context, in SigningKeyCreate) (domain.SigningKey, error)

	// Rotate moves the window of the current key to in.CurrentKeyWindow and creates in.NextKey in one transaction.
	// It returns domain.ErrSigningKeyNotFound when the current key no longer exists.
	Rotate(ctx context.Context, in SigningKeyRotate) (domain.SigningKey, error)

	// DeleteExpired deletes every key whose delete_after is at or before now and returns their key IDs.
	DeleteExpired(ctx context.Context, now time.Time) ([]string, error)
}
