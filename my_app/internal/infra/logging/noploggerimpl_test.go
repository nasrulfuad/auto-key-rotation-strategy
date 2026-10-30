package logging_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
)

func TestNopLoggerDiscardsEveryLevelAndStaysNopAfterWith(t *testing.T) {
	logger := logging.NopLogger{}

	assert.NotPanics(t, func() {
		logger.Debug(context.Background(), "debug", "key", "value")
		logger.Info(context.Background(), "info")
		logger.Warn(context.Background(), "warn")
		logger.Error(context.Background(), "error")
	})
	assert.Equal(t, logging.NopLogger{}, logger.With("key", "value"))
}
