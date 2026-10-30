package peersigningkey_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/jwks/jwksmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/peersigningkey"
)

func TestListPeerSigningKeysReturnsCachedKeysWithFetchStatus(t *testing.T) {
	peerKeySet := jwksmocks.NewMockPeerKeySet(t)
	fetchedAt := time.Date(2026, 1, 1, 0, 0, 10, 0, time.UTC)
	status := domain.PeerKeySetStatus{URL: "http://peer/.well-known/jwks.json", LastFetchedAt: fetchedAt, LastAttemptAt: fetchedAt}
	keys := []domain.PeerSigningKey{{KeyID: "peer-active", Algorithm: domain.SigningAlgorithmRS256, FetchedAt: fetchedAt}}
	peerKeySet.EXPECT().Status().Return(status)
	peerKeySet.EXPECT().ListAll(mock.Anything).Return(keys)

	out := peersigningkey.NewListPeerSigningKeysInteractor(peerKeySet).Execute(context.Background())

	assert.Equal(t, peersigningkey.ListPeerSigningKeysOutput{Status: status, Keys: keys}, out)
}

func TestListPeerSigningKeysReportsLastErrorBeforeFirstFetch(t *testing.T) {
	peerKeySet := jwksmocks.NewMockPeerKeySet(t)
	status := domain.PeerKeySetStatus{URL: "http://peer/.well-known/jwks.json", LastError: "connection refused"}
	peerKeySet.EXPECT().Status().Return(status)
	peerKeySet.EXPECT().ListAll(mock.Anything).Return(nil)

	out := peersigningkey.NewListPeerSigningKeysInteractor(peerKeySet).Execute(context.Background())

	assert.Equal(t, "connection refused", out.Status.LastError)
	assert.Empty(t, out.Keys)
}
