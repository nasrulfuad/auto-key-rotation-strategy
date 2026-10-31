package logging

import (
	"context"
	"io"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

type SlogLogger struct {
	handler slog.Handler
}

var _ Logger = SlogLogger{}

func New(opts ...Option) SlogLogger {
	return SlogLogger{handler: newOptions(opts).handler}
}

func NewJSON(writer io.Writer, opts ...Option) SlogLogger {
	o := newOptions(opts)
	handler := slog.NewJSONHandler(writer, &slog.HandlerOptions{Level: parseLevel(o.level)})
	return New(WithHandler(handler.WithAttrs([]slog.Attr{slog.String("service", o.service)})))
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func (l SlogLogger) Debug(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelDebug, msg, args)
}

func (l SlogLogger) Info(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelInfo, msg, args)
}

func (l SlogLogger) Warn(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelWarn, msg, args)
}

func (l SlogLogger) Error(ctx context.Context, msg string, args ...any) {
	l.write(ctx, slog.LevelError, msg, args)
}

func (l SlogLogger) With(args ...any) Logger {
	return SlogLogger{handler: slog.New(l.handler).With(args...).Handler()}
}

func (l SlogLogger) write(ctx context.Context, level slog.Level, msg string, args []any) {
	if ctx == nil {
		ctx = context.Background()
	}
	if !l.handler.Enabled(ctx, level) {
		return
	}
	var callers [1]uintptr
	runtime.Callers(3, callers[:])
	record := slog.NewRecord(time.Now(), level, msg, callers[0])
	record.Add(ListContextFields(ctx)...)
	record.Add(args...)
	_ = l.handler.Handle(ctx, record)
}
