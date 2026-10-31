package repository

import "context"

// JobLock lets one instance at a time run a scheduler job.
type JobLock interface {
	// WithLock calls run while holding the lock named jobName for the length of a database transaction.
	// It returns false without calling run when another holder has the lock, and otherwise returns true and the error of run.
	WithLock(ctx context.Context, jobName string, run func(ctx context.Context) error) (bool, error)
}
