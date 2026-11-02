package presenter

import (
	"time"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/peersigningkey"
)

type PeerSigningKeyPresenter struct{}

func NewPeerSigningKeyPresenter(opts ...Option) *PeerSigningKeyPresenter {
	return &PeerSigningKeyPresenter{}
}

type peerSigningKeyListBody struct {
	URL           string               `json:"url"`
	LastFetchedAt *time.Time           `json:"last_fetched_at"`
	LastAttemptAt *time.Time           `json:"last_attempt_at"`
	LastError     *string              `json:"last_error"`
	Data          []peerSigningKeyBody `json:"data"`
}

type peerSigningKeyBody struct {
	KeyID     string    `json:"key_id"`
	Algorithm string    `json:"algorithm"`
	FetchedAt time.Time `json:"fetched_at"`
}

func (p *PeerSigningKeyPresenter) ListPeerSigningKeys(out peersigningkey.ListPeerSigningKeysOutput) peerSigningKeyListBody {
	body := peerSigningKeyListBody{
		URL:           out.Status.URL,
		LastFetchedAt: optionalTime(out.Status.LastFetchedAt),
		LastAttemptAt: optionalTime(out.Status.LastAttemptAt),
		Data:          make([]peerSigningKeyBody, 0, len(out.Keys)),
	}
	if out.Status.LastError != "" {
		body.LastError = &out.Status.LastError
	}
	for _, key := range out.Keys {
		body.Data = append(body.Data, peerSigningKeyBody{KeyID: key.KeyID, Algorithm: string(key.Algorithm), FetchedAt: key.FetchedAt})
	}
	return body
}

func optionalTime(value time.Time) *time.Time {
	if value.IsZero() {
		return nil
	}
	return &value
}
