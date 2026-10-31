//go:build integration

package integration_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwks"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwt"
)

const probeAudience domain.TokenAudience = "rotation.probe"

func newSelfVerifier(a *app, refreshOnUnknownKeyID bool) (*jwt.Verifier, *jwks.RemoteKeySet) {
	keySet := jwks.NewRemoteKeySet(a.jwksURL(), jwks.WithRefreshOnUnknownKeyID(refreshOnUnknownKeyID), jwks.WithRefreshMinimumInterval(10*time.Second), jwks.WithClock(a.clock))
	return jwt.NewVerifier(keySet, domain.SigningAlgorithmRS256), keySet
}

type issuedToken struct {
	raw       string
	keyID     string
	expiresAt time.Time
}

func TestFullRotationKeepsTokensValidUntilTheirKeyIsDeleted(t *testing.T) {
	a := newApp(t, nil)
	ctx := context.Background()
	verifier, keySet := newSelfVerifier(a, true)
	expectation := domain.TokenExpectation{Issuer: a.config.TokenIssuer, Audience: probeAudience}

	lastTokenByKeyID := map[string]issuedToken{}
	deleteAfterByKeyID := map[string]time.Time{}
	var signingOrder []string
	end := startTime.Add(75 * time.Minute)
	for ; a.clock.Now().Before(end); a.clock.Advance(15 * time.Second) {
		a.tick(t)
		require.NoError(t, keySet.Refresh(ctx))
		now := a.clock.Now()

		active := 0
		for _, key := range a.signingKeys(t) {
			deleteAfterByKeyID[key.KeyID] = key.DeleteAfter
			if key.Status == string(domain.SigningKeyStatusActive) {
				active++
			}
			assert.True(t, key.DeleteAfter.After(now), "an expired key must not be listed at %s", now)
		}
		require.Equal(t, 1, active, "exactly one active key at %s", now)

		signed, err := a.container.Signer.Sign(ctx, domain.TokenRequest{Audience: probeAudience})
		require.NoError(t, err)
		if len(signingOrder) == 0 || signingOrder[len(signingOrder)-1] != signed.KeyID {
			signingOrder = append(signingOrder, signed.KeyID)
			assert.Equal(t, 0, int(now.Sub(startTime)%(20*time.Minute)), "signer switched keys at %s, not at an active_from boundary", now)
		}
		lastTokenByKeyID[signed.KeyID] = issuedToken{raw: signed.Value, keyID: signed.KeyID, expiresAt: signed.ExpiresAt}

		for keyID, token := range lastTokenByKeyID {
			_, err := verifier.Verify(ctx, token.raw, expectation, now)
			switch {
			case !now.Before(deleteAfterByKeyID[keyID]):
				assert.ErrorIs(t, err, domain.ErrTokenSigningKeyNotFound, "token of deleted key %s at %s", keyID, now)
			case now.Before(token.expiresAt):
				assert.NoError(t, err, "token of key %s must stay valid at %s", keyID, now)
			default:
				assert.ErrorIs(t, err, domain.TokenClaimInvalid("token_expired"))
			}
		}
	}
	assert.Len(t, signingOrder, 4)
}

func TestVerifierWithoutPollingFailsWhenTheNewKeyActivates(t *testing.T) {
	a := newApp(t, nil)
	ctx := context.Background()
	verifier, keySet := newSelfVerifier(a, false)
	polledVerifier, polledKeySet := newSelfVerifier(a, false)
	expectation := domain.TokenExpectation{Issuer: a.config.TokenIssuer, Audience: probeAudience}

	a.tick(t)
	require.NoError(t, keySet.Refresh(ctx))
	for ; a.clock.Now().Before(startTime.Add(20 * time.Minute)); a.clock.Advance(15 * time.Second) {
		a.tick(t)
		if a.clock.Now().Sub(startTime)%time.Minute == 0 {
			require.NoError(t, polledKeySet.Refresh(ctx))
		}
	}
	a.tick(t)

	signed, err := a.container.Signer.Sign(ctx, domain.TokenRequest{Audience: probeAudience})
	require.NoError(t, err)

	_, err = verifier.Verify(ctx, signed.Value, expectation, a.clock.Now())
	assert.ErrorIs(t, err, domain.ErrTokenSigningKeyNotFound, "a verifier that never polled cannot know the new key")

	_, err = polledVerifier.Verify(ctx, signed.Value, expectation, a.clock.Now())
	assert.NoError(t, err, "a verifier that polled during the pending phase already knows the new key")

	require.NoError(t, keySet.Refresh(ctx))
	_, err = verifier.Verify(ctx, signed.Value, expectation, a.clock.Now())
	assert.NoError(t, err)
}

func TestRecoveryAfterDowntimeCreatesActiveKeyImmediately(t *testing.T) {
	a := newApp(t, nil)
	a.tick(t)
	a.clock.Advance(3 * time.Hour)
	a.tick(t)

	keys := a.signingKeys(t)
	require.Len(t, keys, 1)
	assert.Equal(t, string(domain.SigningKeyStatusActive), keys[0].Status)
	assert.Equal(t, a.clock.Now(), keys[0].ActiveFrom)
}
