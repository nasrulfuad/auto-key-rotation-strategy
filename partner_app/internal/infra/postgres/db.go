package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	partnerapp "github.com/nasrulfuad/auto-key-rotation-strategy/partner_app"
)

const minimumConnections = 2

func Open(databaseURL string, maximumConnections int) (*sql.DB, error) {
	connectionConfig, err := pgx.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}
	db := stdlib.OpenDB(*connectionConfig)
	db.SetMaxOpenConns(max(maximumConnections, minimumConnections))
	db.SetMaxIdleConns(max(maximumConnections, minimumConnections))
	db.SetConnMaxLifetime(time.Hour)
	return db, nil
}

func Migrate(db *sql.DB) error {
	migrations, err := fs.Sub(partnerapp.MigrationFS, "migrations")
	if err != nil {
		return err
	}
	provider, err := goose.NewProvider(goose.DialectPostgres, db, migrations)
	if err != nil {
		return fmt.Errorf("failed to build migration provider: %w", err)
	}
	if _, err := provider.Up(context.Background()); err != nil && !errors.Is(err, goose.ErrNoNextVersion) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}
	return nil
}
