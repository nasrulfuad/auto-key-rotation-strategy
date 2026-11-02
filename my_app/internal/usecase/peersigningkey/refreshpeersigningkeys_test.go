package peersigningkey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/jwks/jwksmocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/peersigningkey"
)

func TestRefreshPeerSigningKeysRefreshesTheCachedKeySet(t *testing.T) {
	peerKeySet := jwksmocks.NewMockPeerKeySet(t)
	peerKeySet.EXPECT().Refresh(mock.Anything).Return(nil)

	err := peersigningkey.NewRefreshPeerSigningKeysInteractor(peerKeySet).Execute(context.Background())

	assert.NoError(t, err)
}

func TestRefreshPeerSigningKeysReturnsFetchError(t *testing.T) {
	peerKeySet := jwksmocks.NewMockPeerKeySet(t)
	fetchErr := errors.New("peer returned 503")
	peerKeySet.EXPECT().Refresh(mock.Anything).Return(fetchErr)

	err := peersigningkey.NewRefreshPeerSigningKeysInteractor(peerKeySet).Execute(context.Background())

	assert.ErrorIs(t, err, fetchErr)
}
