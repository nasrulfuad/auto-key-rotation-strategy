package postgres

import (
	"context"
	"database/sql"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/store"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

type JobLockRepository struct {
	db *sql.DB
}

var _ port.JobLock = (*JobLockRepository)(nil)

func NewJobLockRepository(db *sql.DB) *JobLockRepository {
	return &JobLockRepository{db: db}
}

func (r *JobLockRepository) WithLock(ctx context.Context, jobName string, run func(ctx context.Context) error) (bool, error) {
	acquired := false
	err := withTransaction(ctx, r.db, func(queries *store.Queries) error {
		locked, err := queries.LockJob(ctx, jobName)
		if err != nil {
			return domain.DB(err)
		}
		if !locked {
			return nil
		}
		acquired = true
		return run(ctx)
	})
	return acquired, err
}

type DatabaseProbeRepository struct {
	queries store.Querier
}

var _ port.DatabaseProbe = (*DatabaseProbeRepository)(nil)

func NewDatabaseProbeRepository(queries store.Querier) *DatabaseProbeRepository {
	return &DatabaseProbeRepository{queries: queries}
}

func (r *DatabaseProbeRepository) Ping(ctx context.Context) error {
	if _, err := r.queries.Ping(ctx); err != nil {
		return domain.DB(err)
	}
	return nil
}
