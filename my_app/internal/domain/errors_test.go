package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

func TestErrorIsMatchesCodeWithAnyReason(t *testing.T) {
	assert.ErrorIs(t, domain.TokenClaimInvalid("audience_mismatch"), domain.ErrTokenClaimInvalid)
	assert.ErrorIs(t, domain.TokenInvalid("token_missing"), domain.ErrTokenInvalid)
	assert.NotErrorIs(t, domain.TokenInvalid("token_missing"), domain.ErrTokenClaimInvalid)
	assert.ErrorIs(t, domain.ErrTokenSigningKeyNotFound, domain.ErrUnauthorized)
}

func TestSecretNeverPrintsItsValue(t *testing.T) {
	secret := domain.NewSecret([]byte("private"))

	assert.Equal(t, "[REDACTED]", secret.String())
	_, err := secret.MarshalJSON()
	assert.Error(t, err)
	assert.Equal(t, []byte("private"), secret.Reveal())
}
