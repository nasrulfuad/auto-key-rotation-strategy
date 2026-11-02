package postgres

import (
	"database/sql"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/store"
)

type Option func(*options)

// options are the dependencies of the repositories that have a default.
type options struct {
	// querier runs the sqlc queries; default store.New(db).
	querier store.Querier
}

func WithQuerier(querier store.Querier) Option {
	return func(o *options) {
		if querier != nil {
			o.querier = querier
		}
	}
}

func newOptions(db *sql.DB, opts []Option) options {
	o := options{querier: store.New(db)}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
