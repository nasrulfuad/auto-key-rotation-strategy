package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOptionsDefaultsToSlogDefaultHandlerAndInfoLevel(t *testing.T) {
	o := newOptions(nil)

	assert.Equal(t, slog.Default().Handler(), o.handler)
	assert.Equal(t, "info", o.level)
	assert.Empty(t, o.service)
}

func TestNewJSONAppliesLevelAndService(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewJSON(&buffer, WithLevel("warn"), WithService("probe"))

	logger.Info(context.Background(), "dropped")
	logger.Warn(context.Background(), "kept")

	var line map[string]any
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &line))
	assert.Equal(t, "kept", line["msg"])
	assert.Equal(t, "probe", line["service"])
}

func TestNewAppliesHandlerAndIgnoresNil(t *testing.T) {
	handler := slog.NewTextHandler(&bytes.Buffer{}, nil)

	assert.Equal(t, handler, New(WithHandler(handler)).handler)
	assert.Equal(t, slog.Default().Handler(), New(WithHandler(nil)).handler)
}
