package jwks

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
)

// PeerKeySet is the cached copy of the peer JWKS.
type PeerKeySet interface {
	// ListAll returns the cached peer keys, oldest fetched first.
	ListAll(ctx context.Context) []domain.PeerSigningKey

	// Refresh fetches the peer JWKS and replaces the cached keys. Concurrent calls share one fetch.
	// When the fetch fails the cached keys are kept and the error is returned and recorded in Status.
	Refresh(ctx context.Context) error

	// Status returns the JWKS URL, the time of the last successful fetch and of the last attempt, and the last error.
	Status() domain.PeerKeySetStatus
}
