package postgres

import (
	"context"
	"database/sql"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/store"
)

func withTransaction(ctx context.Context, db *sql.DB, run func(queries *store.Queries) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return domain.DB(err)
	}
	if err := run(store.New(tx)); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := tx.Commit(); err != nil {
		return domain.DB(err)
	}
	return nil
}
