package signingkey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type DeleteExpiredSigningKeysInteractor struct {
	signingKeyWriter port.SigningKeyWriter
	clock            port.Clock
	logger           port.Logger
}

func NewDeleteExpiredSigningKeysInteractor(
	signingKeyWriter port.SigningKeyWriter,
	clock port.Clock,
	logger port.Logger,
) *DeleteExpiredSigningKeysInteractor {
	if logger == nil {
		logger = port.NopLogger{}
	}
	return &DeleteExpiredSigningKeysInteractor{signingKeyWriter: signingKeyWriter, clock: clock, logger: logger}
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
