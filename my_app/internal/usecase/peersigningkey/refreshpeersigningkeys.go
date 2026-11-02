package peersigningkey

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/jwks"
)

type RefreshPeerSigningKeysInteractor struct {
	peerKeySet jwks.PeerKeySet
}

func NewRefreshPeerSigningKeysInteractor(peerKeySet jwks.PeerKeySet) *RefreshPeerSigningKeysInteractor {
	return &RefreshPeerSigningKeysInteractor{peerKeySet: peerKeySet}
}

func (i *RefreshPeerSigningKeysInteractor) Execute(ctx context.Context) error {
	return i.peerKeySet.Refresh(ctx)
}
