package signingkey

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/systemclock"
)

type DeleteExpiredSigningKeysInteractor struct {
	signingKeyWriter repository.SigningKeyWriter
	clock            systemclock.Clock
	logger           logging.Logger
}

func NewDeleteExpiredSigningKeysInteractor(signingKeyWriter repository.SigningKeyWriter, opts ...Option) *DeleteExpiredSigningKeysInteractor {
	o := newOptions(opts)
	return &DeleteExpiredSigningKeysInteractor{signingKeyWriter: signingKeyWriter, clock: o.clock, logger: o.logger}
}

func (i *DeleteExpiredSigningKeysInteractor) Execute(ctx context.Context) error {
	now := i.clock.Now()
	keyIDs, err := i.signingKeyWriter.DeleteExpired(ctx, now)
	if err != nil {
		return err
	}
	for _, keyID := range keyIDs {
		i.logger.Info(ctx, "expired signing key deleted; removed from the database and the JWKS, tokens signed by it are now rejected",
			"kid", keyID, "deleted_at", now)
	}
	return nil
}
