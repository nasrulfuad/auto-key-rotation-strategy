package postgres

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/store"
)

type DatabaseProbeRepository struct {
	queries store.Querier
}

var _ repository.DatabaseProbe = (*DatabaseProbeRepository)(nil)

func NewDatabaseProbeRepository(queries store.Querier) *DatabaseProbeRepository {
	return &DatabaseProbeRepository{queries: queries}
}

func (r *DatabaseProbeRepository) Ping(ctx context.Context) error {
	if _, err := r.queries.Ping(ctx); err != nil {
		return domain.DB(err)
	}
	return nil
}
