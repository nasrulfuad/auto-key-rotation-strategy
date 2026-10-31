package logging

import "context"

// Logger writes structured log lines. Fields added to ctx with AddContextFields (source, job, log_id) are written on every line.
type Logger interface {
	// Debug logs msg at debug level with args as alternating keys and values.
	Debug(ctx context.Context, msg string, args ...any)

	// Info logs msg at info level with args as alternating keys and values.
	Info(ctx context.Context, msg string, args ...any)

	// Warn logs msg at warn level with args as alternating keys and values.
	Warn(ctx context.Context, msg string, args ...any)

	// Error logs msg at error level with args as alternating keys and values.
	Error(ctx context.Context, msg string, args ...any)

	// With returns a Logger that adds args to every line it writes.
	With(args ...any) Logger
}
