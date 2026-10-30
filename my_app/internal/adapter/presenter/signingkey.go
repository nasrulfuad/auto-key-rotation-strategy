package presenter

import (
	"encoding/base64"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/signingkey"
)

type SigningKeyPresenter struct{}

func NewSigningKeyPresenter() *SigningKeyPresenter { return &SigningKeyPresenter{} }

type jsonWebKeySetBody struct {
	Keys []jsonWebKeyBody `json:"keys"`
}

type jsonWebKeyBody struct {
	KeyType   string `json:"kty"`
	Exponent  string `json:"e"`
	Use       string `json:"use"`
	KeyID     string `json:"kid"`
	Algorithm string `json:"alg"`
	Modulus   string `json:"n"`
}

type signingKeyListBody struct {
	Now  time.Time        `json:"now"`
	Data []signingKeyBody `json:"data"`
}

type signingKeyBody struct {
	KeyID        string    `json:"key_id"`
	Algorithm    string    `json:"algorithm"`
	Status       string    `json:"status"`
	IsPublished  bool      `json:"is_published"`
	IntroducedAt time.Time `json:"introduced_at"`
	ActiveFrom   time.Time `json:"active_from"`
	ActiveUntil  time.Time `json:"active_until"`
	DeleteAfter  time.Time `json:"delete_after"`
	CreatedAt    time.Time `json:"created_at"`
}

func (p *SigningKeyPresenter) ListPublishedSigningKeys(keys []domain.SigningKey) jsonWebKeySetBody {
	body := jsonWebKeySetBody{Keys: make([]jsonWebKeyBody, 0, len(keys))}
	for _, key := range keys {
		body.Keys = append(body.Keys, jsonWebKeyBody{
			KeyType:   "RSA",
			Exponent:  base64.RawURLEncoding.EncodeToString(key.PublicKey.Exponent),
			Use:       "sig",
			KeyID:     key.KeyID,
			Algorithm: string(key.Algorithm),
			Modulus:   base64.RawURLEncoding.EncodeToString(key.PublicKey.Modulus),
		})
	}
	return body
}

func (p *SigningKeyPresenter) ListSigningKeys(out signingkey.ListSigningKeysOutput) signingKeyListBody {
	body := signingKeyListBody{Now: out.Now, Data: make([]signingKeyBody, 0, len(out.Keys))}
	for _, key := range out.Keys {
		body.Data = append(body.Data, signingKeyBody{
			KeyID:        key.KeyID,
			Algorithm:    string(key.Algorithm),
			Status:       string(key.StatusAt(out.Now)),
			IsPublished:  key.IsPublishedAt(out.Now),
			IntroducedAt: key.Window.IntroducedAt,
			ActiveFrom:   key.Window.ActiveFrom,
			ActiveUntil:  key.Window.ActiveUntil,
			DeleteAfter:  key.Window.DeleteAfter,
			CreatedAt:    key.CreatedAt,
		})
	}
	return body
}
