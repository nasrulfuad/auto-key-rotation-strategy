package scheduler

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
)

const maximumRunDuration = 5 * time.Minute

type Job struct {
	Name      string
	Interval  time.Duration
	Exclusive bool
	Run       func(ctx context.Context) error
}

type Runner struct {
	jobs    []Job
	jobLock repository.JobLock
	logger  logging.Logger
}

func NewRunner(jobLock repository.JobLock, jobs []Job, opts ...Option) *Runner {
	return &Runner{jobs: jobs, jobLock: jobLock, logger: newOptions(opts).logger}
}

func (r *Runner) Start(ctx context.Context) {
	for _, job := range r.jobs {
		if job.Interval <= 0 {
			continue
		}
		go r.loop(ctx, job)
	}
}

func (r *Runner) RunOnce(ctx context.Context, name string) error {
	for _, job := range r.jobs {
		if job.Name == name {
			return r.run(addJobContextFields(ctx, job), job)
		}
	}
	return nil
}

func (r *Runner) loop(ctx context.Context, job Job) {
	ticker := time.NewTicker(job.Interval)
	defer ticker.Stop()
	for {
		r.runOnce(ctx, job)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (r *Runner) runOnce(ctx context.Context, job Job) {
	ctx = addJobContextFields(ctx, job)
	runCtx, cancel := context.WithTimeout(ctx, min(job.Interval, maximumRunDuration))
	defer cancel()
	startedAt := time.Now()
	if err := r.run(runCtx, job); err != nil {
		if ctx.Err() != nil {
			return
		}
		r.logger.Warn(ctx, "scheduled job failed", "duration_ms", time.Since(startedAt).Milliseconds(), "error", err)
		return
	}
	r.logger.Debug(ctx, "scheduled job completed", "duration_ms", time.Since(startedAt).Milliseconds())
}

func (r *Runner) run(ctx context.Context, job Job) error {
	if !job.Exclusive || r.jobLock == nil {
		return job.Run(ctx)
	}
	acquired, err := r.jobLock.WithLock(ctx, job.Name, job.Run)
	if err != nil {
		return err
	}
	if !acquired {
		r.logger.Debug(ctx, "scheduled job skipped; another instance holds the lock")
	}
	return nil
}

func addJobContextFields(ctx context.Context, job Job) context.Context {
	return logging.AddContextFields(ctx,
		logging.FieldSource, logging.SourceTick, logging.FieldJob, job.Name, logging.FieldLogID, uuid.NewString())
}
