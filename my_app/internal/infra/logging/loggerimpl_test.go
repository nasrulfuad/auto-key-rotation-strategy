package logging_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
)

func decodeLogLine(t *testing.T, buffer *bytes.Buffer) map[string]any {
	t.Helper()
	var line map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &line))
	return line
}

func TestLoggerWritesContextFieldsBeforeCallFields(t *testing.T) {
	var buffer bytes.Buffer
	logger := logging.NewJSON(&buffer, "debug", "my_app")
	ctx := logging.AddContextFields(context.Background(), logging.FieldSource, logging.SourceAPI, logging.FieldLogID, "log-1")

	logger.Info(ctx, "request completed", "status", 200)

	line := decodeLogLine(t, &buffer)
	assert.Equal(t, "api", line["source"])
	assert.Equal(t, "log-1", line["log_id"])
	assert.Equal(t, float64(200), line["status"])
	assert.Equal(t, "my_app", line["service"])
}

func TestLoggerOmitsContextFieldsWhenNoneAreSet(t *testing.T) {
	var buffer bytes.Buffer
	logger := logging.NewJSON(&buffer, "debug", "my_app")

	logger.Info(context.Background(), "signing key created")

	line := decodeLogLine(t, &buffer)
	assert.NotContains(t, line, "log_id")
	assert.NotContains(t, line, "source")
}

func TestAddContextFieldsDoesNotLeakIntoParentContext(t *testing.T) {
	parent := logging.AddContextFields(context.Background(), logging.FieldSource, logging.SourceTick)
	child := logging.AddContextFields(parent, logging.FieldJob, "ensure_signing_keys")

	assert.Equal(t, []any{"source", "tick"}, logging.ListContextFields(parent))
	assert.Equal(t, []any{"source", "tick", "job", "ensure_signing_keys"}, logging.ListContextFields(child))
}
