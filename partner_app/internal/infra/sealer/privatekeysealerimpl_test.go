package sealer_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/sealer"
)

func TestAESGCMRoundTripAndTamperDetection(t *testing.T) {
	s, err := sealer.NewAESGCM(bytes.Repeat([]byte{7}, sealer.KeySize))
	require.NoError(t, err)

	ciphertext, nonce, err := s.Seal([]byte("private key"))
	require.NoError(t, err)
	plaintext, err := s.Open(ciphertext, nonce)
	require.NoError(t, err)
	assert.Equal(t, []byte("private key"), plaintext)

	ciphertext[0] ^= 0xff
	_, err = s.Open(ciphertext, nonce)
	assert.Error(t, err)
}

func TestNewAESGCMRejectsWrongKeySize(t *testing.T) {
	_, err := sealer.NewAESGCM([]byte("short"))
	assert.Error(t, err)
}
