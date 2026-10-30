package repository

import "context"

type JobLock interface {
	WithLock(ctx context.Context, jobName string, run func(ctx context.Context) error) (bool, error)
}
