package port

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type UUIDGenerator interface {
	NewUUID() uuid.UUID
}

type Logger interface {
	Debug(ctx context.Context, msg string, args ...any)
	Info(ctx context.Context, msg string, args ...any)
	Warn(ctx context.Context, msg string, args ...any)
	Error(ctx context.Context, msg string, args ...any)
	With(args ...any) Logger
}

type JobLock interface {
	WithLock(ctx context.Context, jobName string, run func(ctx context.Context) error) (bool, error)
}

type DatabaseProbe interface {
	Ping(ctx context.Context) error
}
