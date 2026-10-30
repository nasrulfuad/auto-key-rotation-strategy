package sealer

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

const KeySize = 32

type AESGCM struct {
	aead cipher.AEAD
}

var _ port.PrivateKeySealer = AESGCM{}

func NewAESGCM(key []byte) (AESGCM, error) {
	if len(key) != KeySize {
		return AESGCM{}, fmt.Errorf("encryption key must be %d bytes, got %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return AESGCM{}, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return AESGCM{}, err
	}
	return AESGCM{aead: aead}, nil
}

func (s AESGCM) Seal(plaintext []byte) ([]byte, []byte, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, nil, err
	}
	return s.aead.Seal(nil, nonce, plaintext, nil), nonce, nil
}

func (s AESGCM) Open(ciphertext, nonce []byte) ([]byte, error) {
	if len(nonce) != s.aead.NonceSize() {
		return nil, fmt.Errorf("nonce must be %d bytes, got %d", s.aead.NonceSize(), len(nonce))
	}
	return s.aead.Open(nil, nonce, ciphertext, nil)
}
