package jwks

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

	assert.Equal(t, 5*time.Second, o.httpClient.Timeout)
	assert.True(t, o.refreshOnUnknownKeyID)
	assert.Equal(t, 10*time.Second, o.refreshMinimumInterval)
	assert.Equal(t, systemclock.New(), o.clock)
	assert.Equal(t, logging.NopLogger{}, o.logger)
}

func TestNewRemoteKeySetAppliesEveryOption(t *testing.T) {
	client := &http.Client{Timeout: time.Minute}
	logger := logging.NewJSON(nil)

	keySet := NewRemoteKeySet("http://peer/.well-known/jwks.json",
		WithHTTPClient(client),
		WithRefreshOnUnknownKeyID(false),
		WithRefreshMinimumInterval(time.Minute),
		WithClock(fixedClock{}),
		WithLogger(logger),
	)

	assert.Equal(t, "http://peer/.well-known/jwks.json", keySet.url)
	assert.Same(t, client, keySet.options.httpClient)
	assert.False(t, keySet.options.refreshOnUnknownKeyID)
	assert.Equal(t, time.Minute, keySet.options.refreshMinimumInterval)
	assert.Equal(t, fixedClock{}, keySet.clock)
	assert.Equal(t, logger, keySet.logger)
}
