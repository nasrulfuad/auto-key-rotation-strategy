package postgres

import (
	"context"
	"database/sql"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/store"
)

type DatabaseProbeRepository struct {
	queries store.Querier
}

var _ repository.DatabaseProbe = (*DatabaseProbeRepository)(nil)

func NewDatabaseProbeRepository(db *sql.DB, opts ...Option) *DatabaseProbeRepository {
	return &DatabaseProbeRepository{queries: newOptions(db, opts).querier}
}

func (r *DatabaseProbeRepository) Ping(ctx context.Context) error {
	if _, err := r.queries.Ping(ctx); err != nil {
		return domain.DB(err)
	}
	return nil
}
