package scheduler_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/scheduler"
)

func TestRunnerTagsEveryRunWithSourceJobAndLogID(t *testing.T) {
	var fieldsPerRun [][]any
	job := scheduler.Job{Name: "ensure_signing_keys", Run: func(ctx context.Context) error {
		fieldsPerRun = append(fieldsPerRun, logging.ListContextFields(ctx))
		return nil
	}}
	runner := scheduler.NewRunner(nil, []scheduler.Job{job})

	require.NoError(t, runner.RunOnce(context.Background(), "ensure_signing_keys"))
	require.NoError(t, runner.RunOnce(context.Background(), "ensure_signing_keys"))

	require.Len(t, fieldsPerRun, 2)
	for _, fields := range fieldsPerRun {
		require.Len(t, fields, 6)
		assert.Equal(t, []any{"source", "tick", "job", "ensure_signing_keys", "log_id"}, fields[:5])
		assert.NotEmpty(t, fields[5])
	}
	assert.NotEqual(t, fieldsPerRun[0][5], fieldsPerRun[1][5])
}
