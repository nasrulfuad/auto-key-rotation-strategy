//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwks"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/jwt"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/rsakey"
)

type fakePartnerApp struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
	keyID      string
	verify     func(token string, audience domain.TokenAudience) error
}

func newFakePartnerApp(t *testing.T) *fakePartnerApp {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	public := rsakey.ToDomainPublicKey(&privateKey.PublicKey)
	keyID, err := rsakey.Thumbprint(public)
	require.NoError(t, err)
	f := &fakePartnerApp{privateKey: privateKey, keyID: keyID}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/jwks.json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "use": "sig", "alg": "RS256", "kid": keyID,
			"n": rsakey.EncodeBase64URL(public.Modulus), "e": rsakey.EncodeBase64URL(public.Exponent),
		}}})
	})
	mux.HandleFunc("PUT /subjects/{subject_id}/entitlements/{entitlement_id}", func(w http.ResponseWriter, r *http.Request) {
		if err := f.verify(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), "set.entitlement.partner-app.local"); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"TOKEN_SIGNING_KEY_NOT_FOUND","message":"authentication required"}}`))
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider": "my_app", "subject_id": r.PathValue("subject_id"), "entitlement_id": r.PathValue("entitlement_id"),
			"products": body["products"], "offer": map[string]string{"free_trial": "UNKNOWN"}, "status": "ACTIVE",
		})
	})
	mux.HandleFunc("POST /subjects/{subject_id}/entitlements/{entitlement_id}/terminate", func(w http.ResponseWriter, r *http.Request) {
		if err := f.verify(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), "set.entitlement.partner-app.local"); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"error":{"code":"TOKEN_SIGNING_KEY_NOT_FOUND","message":"authentication required"}}`))
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"provider": "my_app", "subject_id": r.PathValue("subject_id"), "entitlement_id": r.PathValue("entitlement_id"),
			"products": "com.partner.product1", "offer": map[string]string{"free_trial": "UNKNOWN"},
			"status": "TERMINATED", "terminated_at": "2026-01-01T00:44:00Z",
		})
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakePartnerApp) notificationToken(t *testing.T, subject string, issuedAt time.Time) string {
	raw, err := jwt.Sign(domain.TokenClaims{
		ID: "jti", Issuer: "qa.partner-app.local", Subject: subject, Provider: "my_app",
		Audience: []string{"activation-notification.my-app.local"}, IssuedAt: issuedAt, ExpiresAt: issuedAt.Add(5 * time.Minute),
	}, f.keyID, domain.SigningAlgorithmRS256, f.privateKey)
	require.NoError(t, err)
	return raw
}

func postNotification(t *testing.T, url, token string, body map[string]string) int {
	encoded, err := json.Marshal(body)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(encoded))
	require.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer "+token)
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	_ = response.Body.Close()
	return response.StatusCode
}

func TestActivationAndTerminationAreSignedWithTheActiveKeyAcrossRotation(t *testing.T) {
	partner := newFakePartnerApp(t)
	a := newApp(t, map[string]string{
		"PARTNER_BASE_URL": partner.server.URL,
		"PEER_JWKS_URL":    partner.server.URL + "/.well-known/jwks.json",
	})
	keySet := jwks.NewRemoteKeySet(jwks.Options{URL: a.jwksURL(), RefreshOnUnknownKeyID: true, RefreshMinimumInterval: 10 * time.Second}, a.clock, nil)
	verifier := jwt.NewVerifier(keySet, domain.SigningAlgorithmRS256, nil)
	partner.verify = func(token string, audience domain.TokenAudience) error {
		_, err := verifier.Verify(context.Background(), token, domain.TokenExpectation{Issuer: "qa.my-app.local", Audience: audience}, a.clock.Now())
		return err
	}

	seen := map[string]bool{}
	for ; a.clock.Now().Before(startTime.Add(45 * time.Minute)); a.clock.Advance(time.Minute) {
		a.tick(t)
		body := postSubscription(t, a.server.URL+"/activations", `{"subject_id":"SUB001","entitlement_id":"ENT001","products":"com.partner.product1"}`)
		assert.Equal(t, "ACTIVE", body.Status)
		seen[body.SignedWithKeyID] = true
	}
	assert.Len(t, seen, 3)

	terminated := postSubscription(t, a.server.URL+"/terminations", `{"subject_id":"SUB001","entitlement_id":"ENT001"}`)
	assert.Equal(t, "TERMINATED", terminated.Status)
	require.NotNil(t, terminated.TerminatedAt)
	require.NotNil(t, terminated.ActivatedAt, "termination must keep the local activation time")

	response, err := http.Get(a.server.URL + "/subscriptions")
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	var listed struct {
		Data []subscriptionView `json:"data"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&listed))
	require.Len(t, listed.Data, 1)
	assert.Equal(t, "TERMINATED", listed.Data[0].Status)
}

type subscriptionView struct {
	Status          string     `json:"status"`
	SignedWithKeyID string     `json:"signed_with_key_id"`
	ActivatedAt     *time.Time `json:"activated_at"`
	TerminatedAt    *time.Time `json:"terminated_at"`
}

func postSubscription(t *testing.T, url, body string) subscriptionView {
	t.Helper()
	response, err := http.Post(url, "application/json", strings.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	require.Equal(t, http.StatusOK, response.StatusCode)
	var view subscriptionView
	require.NoError(t, json.NewDecoder(response.Body).Decode(&view))
	return view
}

func TestActivationNotificationWebhookVerifiesPartnerToken(t *testing.T) {
	partner := newFakePartnerApp(t)
	a := newApp(t, map[string]string{"PEER_JWKS_URL": partner.server.URL + "/.well-known/jwks.json"})
	a.tick(t)
	webhook := a.server.URL + "/activation-notifications"
	body := map[string]string{
		"notification_id": uuid.NewString(), "event": "entitlement_upserted", "provider": "my_app",
		"subject_id": "SUB001", "entitlement_id": "ENT001", "products": "p", "free_trial": "NONE",
	}
	token := partner.notificationToken(t, "SUB001", a.clock.Now())

	assert.Equal(t, http.StatusNoContent, postNotification(t, webhook, token, body))
	assert.Equal(t, http.StatusNoContent, postNotification(t, webhook, token, body))
	assert.Equal(t, http.StatusUnauthorized, postNotification(t, webhook, partner.notificationToken(t, "SUB999", a.clock.Now()), body))
	assert.Equal(t, http.StatusUnauthorized, postNotification(t, webhook, "garbage", body))

	response, err := http.Get(webhook)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	var listed struct {
		Data []map[string]any `json:"data"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&listed))
	require.Len(t, listed.Data, 1)
	assert.Equal(t, partner.keyID, listed.Data[0]["key_id"])
}
