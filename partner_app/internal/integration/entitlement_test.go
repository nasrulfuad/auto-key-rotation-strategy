//go:build integration

package integration_test

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/di"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/jwks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/jwt"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/rsakey"
)

type fakeMyApp struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
	keyID      string

	mu            sync.Mutex
	notifications []receivedNotification
}

type receivedNotification struct {
	token string
	body  map[string]string
}

func newFakeMyApp(t *testing.T) *fakeMyApp {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	public := rsakey.ToDomainPublicKey(&privateKey.PublicKey)
	keyID, err := rsakey.Thumbprint(public)
	require.NoError(t, err)
	f := &fakeMyApp{privateKey: privateKey, keyID: keyID}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /.well-known/jwks.json", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]string{{
			"kty": "RSA", "use": "sig", "alg": "RS256", "kid": keyID,
			"n": rsakey.EncodeBase64URL(public.Modulus), "e": rsakey.EncodeBase64URL(public.Exponent),
		}}})
	})
	mux.HandleFunc("POST /activation-notifications", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		f.mu.Lock()
		f.notifications = append(f.notifications, receivedNotification{token: strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), body: body})
		f.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	})
	f.server = httptest.NewServer(mux)
	t.Cleanup(f.server.Close)
	return f
}

func (f *fakeMyApp) token(t *testing.T, audience, provider string, issuedAt time.Time) string {
	raw, err := jwt.Sign(domain.TokenClaims{
		ID: "jti", Issuer: "qa.my-app.local", Provider: provider, Audience: []string{audience},
		IssuedAt: issuedAt, ExpiresAt: issuedAt.Add(5 * time.Minute),
	}, f.keyID, domain.SigningAlgorithmRS256, f.privateKey)
	require.NoError(t, err)
	return raw
}

type apiResponse struct {
	status int
	body   map[string]any
}

func call(t *testing.T, method, url, token, body string) apiResponse {
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	request, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	var decoded map[string]any
	_ = json.NewDecoder(response.Body).Decode(&decoded)
	return apiResponse{status: response.StatusCode, body: decoded}
}

func errorCode(r apiResponse) string {
	envelope, _ := r.body["error"].(map[string]any)
	code, _ := envelope["code"].(string)
	return code
}

func TestEntitlementLifecycleWithPeerTokens(t *testing.T) {
	peer := newFakeMyApp(t)
	a := newApp(t, map[string]string{
		"PEER_JWKS_URL":               peer.server.URL + "/.well-known/jwks.json",
		"ACTIVATION_NOTIFICATION_URL": peer.server.URL + "/activation-notifications",
	})
	a.tick(t)
	now := a.clock.Now()
	entitlementURL := a.server.URL + "/subjects/SUB001/entitlements/ENT001"
	setToken := peer.token(t, "set.entitlement.partner-app.local", "my_app", now)
	getToken := peer.token(t, "get.entitlement.partner-app.local", "my_app", now)

	put := call(t, http.MethodPut, entitlementURL, setToken, `{"products":"com.partner.product1","offer":{"free_trial":"NONE"}}`)
	require.Equal(t, http.StatusOK, put.status, put.body)
	assert.Equal(t, "com.partner.product1", put.body["products"])

	get := call(t, http.MethodGet, entitlementURL, getToken, "")
	require.Equal(t, http.StatusOK, get.status)

	list := call(t, http.MethodGet, a.server.URL+"/subjects/SUB001/entitlements", getToken, "")
	require.Equal(t, http.StatusOK, list.status)
	assert.Len(t, list.body["data"], 1)

	wrongAudience := call(t, http.MethodGet, entitlementURL, setToken, "")
	assert.Equal(t, http.StatusUnauthorized, wrongAudience.status)
	assert.Equal(t, "TOKEN_CLAIM_INVALID", errorCode(wrongAudience))

	stale := call(t, http.MethodPut, entitlementURL, peer.token(t, "set.entitlement.partner-app.local", "my_app", now.Add(-time.Minute)), `{"products":"older"}`)
	assert.Equal(t, http.StatusBadRequest, stale.status)
	assert.Equal(t, "ENTITLEMENT_STALE_UPDATE", errorCode(stale))

	foreign := call(t, http.MethodPut, entitlementURL, peer.token(t, "set.entitlement.partner-app.local", "someone_else", now), `{"products":"x"}`)
	assert.Equal(t, "TOKEN_CLAIM_INVALID", errorCode(foreign))

	invalidSubject := call(t, http.MethodPut, a.server.URL+"/subjects/nil/entitlements/ENT001", setToken, `{"products":"x"}`)
	assert.Equal(t, "SUBJECT_ID_INVALID", errorCode(invalidSubject))

	missing := call(t, http.MethodGet, a.server.URL+"/subjects/SUB001/entitlements/ENT999", getToken, "")
	assert.Equal(t, http.StatusNotFound, missing.status)
	assert.Equal(t, "ENTITLEMENT_NOT_FOUND", errorCode(missing))

	unsigned := call(t, http.MethodGet, entitlementURL, "", "")
	assert.Equal(t, "TOKEN_INVALID", errorCode(unsigned))

	require.NoError(t, a.container.Scheduler.RunOnce(context.Background(), di.JobDeliverActivationNotifications))
	peer.mu.Lock()
	notifications := append([]receivedNotification(nil), peer.notifications...)
	peer.mu.Unlock()
	require.Len(t, notifications, 1)
	assert.Equal(t, "SUB001", notifications[0].body["subject_id"])

	terminateURL := entitlementURL + "/terminate"
	missingTermination := call(t, http.MethodPost, a.server.URL+"/subjects/SUB001/entitlements/ENT999/terminate", setToken, "")
	assert.Equal(t, "ENTITLEMENT_NOT_FOUND", errorCode(missingTermination))
	readOnlyTermination := call(t, http.MethodPost, terminateURL, getToken, "")
	assert.Equal(t, "TOKEN_CLAIM_INVALID", errorCode(readOnlyTermination))
	staleTermination := call(t, http.MethodPost, terminateURL, peer.token(t, "set.entitlement.partner-app.local", "my_app", now.Add(-time.Minute)), "")
	assert.Equal(t, "ENTITLEMENT_STALE_UPDATE", errorCode(staleTermination))

	a.clock.Advance(time.Minute)
	terminated := call(t, http.MethodPost, terminateURL, peer.token(t, "set.entitlement.partner-app.local", "my_app", a.clock.Now()), "")
	require.Equal(t, http.StatusOK, terminated.status, terminated.body)
	assert.Equal(t, "TERMINATED", terminated.body["status"])
	assert.NotNil(t, terminated.body["terminated_at"])

	afterTermination := call(t, http.MethodGet, entitlementURL, peer.token(t, "get.entitlement.partner-app.local", "my_app", a.clock.Now()), "")
	assert.Equal(t, "TERMINATED", afterTermination.body["status"])

	a.clock.Advance(time.Minute)
	reactivated := call(t, http.MethodPut, entitlementURL, peer.token(t, "set.entitlement.partner-app.local", "my_app", a.clock.Now()), `{"products":"com.partner.product1"}`)
	require.Equal(t, http.StatusOK, reactivated.status)
	assert.Equal(t, "ACTIVE", reactivated.body["status"])
	assert.Nil(t, reactivated.body["terminated_at"])

	require.NoError(t, a.container.Scheduler.RunOnce(context.Background(), di.JobDeliverActivationNotifications))
	peer.mu.Lock()
	events := []string{}
	for _, notification := range peer.notifications {
		events = append(events, notification.body["event"])
	}
	peer.mu.Unlock()
	assert.Equal(t, []string{"entitlement_upserted", "entitlement_terminated", "entitlement_upserted"}, events)

	keySet := jwks.NewRemoteKeySet(a.jwksURL(), jwks.WithRefreshOnUnknownKeyID(true), jwks.WithClock(a.clock))
	verified, err := jwt.NewVerifier(keySet, domain.SigningAlgorithmRS256).Verify(context.Background(), notifications[0].token,
		domain.TokenExpectation{Issuer: "qa.partner-app.local", Audience: "activation-notification.my-app.local"}, a.clock.Now())
	require.NoError(t, err)
	assert.Equal(t, "SUB001", verified.Claims.Subject)

	require.NoError(t, a.container.Scheduler.RunOnce(context.Background(), di.JobDeliverActivationNotifications))
	peer.mu.Lock()
	assert.Len(t, peer.notifications, 3, "a sent notification must not be delivered twice")
	peer.mu.Unlock()
}
