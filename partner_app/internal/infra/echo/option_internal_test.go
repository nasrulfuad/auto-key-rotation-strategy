package echo

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
)

func TestNewOptionsDefaultsToFifteenSecondTimeoutAndNopLogger(t *testing.T) {
	o := newOptions(nil)

	assert.Equal(t, 15*time.Second, o.requestTimeout)
	assert.Equal(t, logging.NopLogger{}, o.logger)
}

func TestNewOptionsAppliesRequestTimeoutAndLogger(t *testing.T) {
	logger := logging.NewJSON(nil)

	o := newOptions([]Option{WithRequestTimeout(time.Second), WithLogger(logger), WithLogger(nil)})

	assert.Equal(t, time.Second, o.requestTimeout)
	assert.Equal(t, logger, o.logger)
}
