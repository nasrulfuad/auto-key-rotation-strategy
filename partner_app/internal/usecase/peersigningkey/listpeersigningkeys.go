package peersigningkey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type ListPeerSigningKeysOutput struct {
	Status domain.PeerKeySetStatus
	Keys   []domain.PeerSigningKey
}

type ListPeerSigningKeysInteractor struct {
	peerKeySet port.PeerKeySet
}

func NewListPeerSigningKeysInteractor(peerKeySet port.PeerKeySet) *ListPeerSigningKeysInteractor {
	return &ListPeerSigningKeysInteractor{peerKeySet: peerKeySet}
}

func (i *ListPeerSigningKeysInteractor) Execute(ctx context.Context) ListPeerSigningKeysOutput {
	return ListPeerSigningKeysOutput{Status: i.peerKeySet.Status(), Keys: i.peerKeySet.ListAll(ctx)}
}
