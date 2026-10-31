package di

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

func TestNewOptionsDefaults(t *testing.T) {
	o := newOptions(nil)

	assert.Equal(t, logging.NopLogger{}, o.logger)
	assert.Equal(t, systemclock.New(), o.clock)
	assert.Equal(t, 5*time.Second, o.httpClient.Timeout)
	assert.Empty(t, o.openAPISpec)
}

func TestNewOptionsAppliesEveryOption(t *testing.T) {
	logger := logging.NewJSON(nil)
	client := &http.Client{Timeout: time.Minute}

	o := newOptions([]Option{WithLogger(logger), WithClock(fixedClock{}), WithHTTPClient(client), WithOpenAPISpec([]byte("{}"))})

	assert.Equal(t, logger, o.logger)
	assert.Equal(t, fixedClock{}, o.clock)
	assert.Same(t, client, o.httpClient)
	assert.Equal(t, []byte("{}"), o.openAPISpec)
}
