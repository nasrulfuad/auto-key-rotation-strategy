package repository

import "context"

// DatabaseProbe checks that the database answers queries.
type DatabaseProbe interface {
	// Ping runs a trivial query and returns a DB error when the database cannot answer.
	Ping(ctx context.Context) error
}
