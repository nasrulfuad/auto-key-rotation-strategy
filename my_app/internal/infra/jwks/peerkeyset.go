package jwks

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type PeerKeySet interface {
	ListAll(ctx context.Context) []domain.PeerSigningKey
	Refresh(ctx context.Context) error
	Status() domain.PeerKeySetStatus
}
