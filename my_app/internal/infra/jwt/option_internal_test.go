package jwt

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
)

func TestNewOptionsDefaultsToNopLogger(t *testing.T) {
	assert.Equal(t, logging.NopLogger{}, newOptions(nil).logger)
}

func TestNewOptionsAppliesLoggerAndIgnoresNil(t *testing.T) {
	logger := logging.NewJSON(nil)

	assert.Equal(t, logger, newOptions([]Option{WithLogger(logger)}).logger)
	assert.Equal(t, logging.NopLogger{}, newOptions([]Option{WithLogger(nil)}).logger)
}
