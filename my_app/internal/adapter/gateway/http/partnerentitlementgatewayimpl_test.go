package http_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/gateway"
	httpgateway "github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/gateway/http"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
)

func TestPartnerEntitlementClientSendsBearerAndMapsRejections(t *testing.T) {
	var authorization, method, path string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		method, path = r.Method, r.URL.Path
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":{"code":"TOKEN_SIGNING_KEY_NOT_FOUND","message":"authentication required"}}`))
	}))
	defer server.Close()

	_, err := httpgateway.NewPartnerEntitlementClient(server.URL).Terminate(context.Background(),
		gateway.PartnerEntitlementTerminate{Token: "jwt", SubjectID: "SUB1", EntitlementID: "ENT1"})

	assert.Equal(t, "Bearer jwt", authorization)
	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/subjects/SUB1/entitlements/ENT1/terminate", path)
	var appErr *domain.Error
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, "PARTNER_REQUEST_REJECTED", appErr.Code())
	assert.Contains(t, appErr.Message(), "401 TOKEN_SIGNING_KEY_NOT_FOUND")
}

func TestPartnerEntitlementClientReportsUnreachablePartner(t *testing.T) {
	server := httptest.NewServer(http.NotFoundHandler())
	server.Close()

	_, err := httpgateway.NewPartnerEntitlementClient(server.URL).Upsert(context.Background(), gateway.PartnerEntitlementUpsert{Token: "jwt"})

	assert.ErrorIs(t, err, domain.ErrPartnerUnavailable)
}
