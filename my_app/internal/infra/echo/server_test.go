package echo_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	controller "github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/controller/http"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/presenter"
	echoserver "github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/echo"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
)

func TestServerTagsRequestContextAndResponseWithLogID(t *testing.T) {
	var buffer bytes.Buffer
	server := echoserver.New(echoserver.Options{RequestTimeout: time.Second, Logger: logging.NewJSON(&buffer, "debug", "test")}, controller.NewRouter(nil, nil, nil, nil, nil, nil, nil, controller.Audiences{ActivationNotification: "activation-notification.my-app.local"}),
		presenter.NewErrorPresenter(nil))
	var fields []any
	server.GET("/probe", func(ectx echo.Context) error {
		fields = logging.ListContextFields(ectx.Request().Context())
		return ectx.NoContent(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()

	server.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/probe?verbose=true", nil))

	logID := recorder.Header().Get(echoserver.HeaderLogID)
	require.NotEmpty(t, logID)
	assert.Equal(t, []any{"source", "api", "log_id", logID}, fields)
	var line map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &line))
	assert.Equal(t, "request completed", line["msg"])
	assert.Equal(t, "/probe?verbose=true", line["uri"])
	assert.Equal(t, "/probe", line["path"])
	assert.Equal(t, logID, line["log_id"])
}
