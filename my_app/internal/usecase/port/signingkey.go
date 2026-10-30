package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
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

type SigningKeyReader interface {
	FindLatest(ctx context.Context) (*domain.SigningKey, error)
	FindActive(ctx context.Context, now time.Time) (domain.SigningKey, error)
	FindPrivateKey(ctx context.Context, keyID string) (domain.Secret, error)
	ListAllPublished(ctx context.Context, now time.Time) ([]domain.SigningKey, error)
	ListAll(ctx context.Context) ([]domain.SigningKey, error)
}

type SigningKeyWriter interface {
	Create(ctx context.Context, in SigningKeyCreate) (domain.SigningKey, error)
	Rotate(ctx context.Context, in SigningKeyRotate) (domain.SigningKey, error)
	DeleteExpired(ctx context.Context, now time.Time) ([]string, error)
}

type KeyPairGenerator interface {
	Generate(ctx context.Context, algorithm domain.SigningAlgorithm) (domain.KeyPair, error)
}

type PrivateKeySealer interface {
	Seal(plaintext []byte) (ciphertext, nonce []byte, err error)
	Open(ciphertext, nonce []byte) ([]byte, error)
}
