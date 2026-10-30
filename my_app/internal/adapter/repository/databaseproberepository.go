package repository

import "context"

type DatabaseProbe interface {
	Ping(ctx context.Context) error
}
