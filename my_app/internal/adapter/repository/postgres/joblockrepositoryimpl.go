package postgres

import (
	"context"
	"database/sql"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/store"
)

type JobLockRepository struct {
	db *sql.DB
}

var _ repository.JobLock = (*JobLockRepository)(nil)

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
