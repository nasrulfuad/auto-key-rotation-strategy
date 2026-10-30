//go:build integration

package integration_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

func (a *app) rotateSigningKey(t *testing.T) (int, map[string]any) {
	t.Helper()
	response, err := http.Post(a.server.URL+"/signing-keys/rotate", "application/json", nil)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	var body map[string]any
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	return response.StatusCode, body
}

func TestManualRotationIntroducesPendingKeyAndSwitchesSignerAfterPendingPhase(t *testing.T) {
	a := newApp(t, nil)
	ctx := context.Background()
	verifier, keySet := newSelfVerifier(a, true)
	expectation := domain.TokenExpectation{Issuer: a.config.TokenIssuer, Audience: probeAudience}
	a.tick(t)
	a.clock.Advance(3 * time.Minute)
	rotatedAt := a.clock.Now()

	status, body := a.rotateSigningKey(t)

	require.Equal(t, http.StatusCreated, status)
	nextKeyID := body["next_key"].(map[string]any)["key_id"].(string)
	currentKeyID := body["current_key"].(map[string]any)["key_id"].(string)
	keys := a.signingKeys(t)
	require.Len(t, keys, 2)
	assert.Equal(t, currentKeyID, keys[0].KeyID)
	assert.Equal(t, string(domain.SigningKeyStatusActive), keys[0].Status)
	assert.Equal(t, rotatedAt.Add(5*time.Minute), keys[0].ActiveUntil)
	assert.Equal(t, rotatedAt.Add(15*time.Minute), keys[0].DeleteAfter)
	assert.Equal(t, nextKeyID, keys[1].KeyID)
	assert.Equal(t, string(domain.SigningKeyStatusPending), keys[1].Status)
	assert.Equal(t, keys[0].ActiveUntil, keys[1].ActiveFrom)

	status, body = a.rotateSigningKey(t)

	assert.Equal(t, http.StatusConflict, status)
	assert.Equal(t, "SIGNING_KEY_ALREADY_PENDING", body["error"].(map[string]any)["code"])

	oldToken, err := a.container.Signer.Sign(ctx, domain.TokenRequest{Audience: probeAudience})
	require.NoError(t, err)
	assert.Equal(t, currentKeyID, oldToken.KeyID)

	a.clock.Advance(5 * time.Minute)
	a.tick(t)
	require.NoError(t, keySet.Refresh(ctx))
	newToken, err := a.container.Signer.Sign(ctx, domain.TokenRequest{Audience: probeAudience})
	require.NoError(t, err)
	assert.Equal(t, nextKeyID, newToken.KeyID)
	_, err = verifier.Verify(ctx, oldToken.Value, expectation, a.clock.Now())
	assert.NoError(t, err)
	_, err = verifier.Verify(ctx, newToken.Value, expectation, a.clock.Now())
	assert.NoError(t, err)

	status, _ = a.rotateSigningKey(t)

	assert.Equal(t, http.StatusCreated, status)
}
