package peersigningkey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type RefreshPeerSigningKeysInteractor struct {
	peerKeySet port.PeerKeySet
}

func NewRefreshPeerSigningKeysInteractor(peerKeySet port.PeerKeySet) *RefreshPeerSigningKeysInteractor {
	return &RefreshPeerSigningKeysInteractor{peerKeySet: peerKeySet}
}

func (i *RefreshPeerSigningKeysInteractor) Execute(ctx context.Context) error {
	return i.peerKeySet.Refresh(ctx)
}
