package signingkey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type EnsureSigningKeysInteractor struct {
	signingKeyReader port.SigningKeyReader
	signingKeyWriter port.SigningKeyWriter
	keyPairGenerator port.KeyPairGenerator
	clock            port.Clock
	uuidGenerator    port.UUIDGenerator
	logger           port.Logger

	schedule  domain.SigningKeySchedule
	algorithm domain.SigningAlgorithm
}

func NewEnsureSigningKeysInteractor(
	signingKeyReader port.SigningKeyReader,
	signingKeyWriter port.SigningKeyWriter,
	keyPairGenerator port.KeyPairGenerator,
	clock port.Clock,
	uuidGenerator port.UUIDGenerator,
	logger port.Logger,
	schedule domain.SigningKeySchedule,
	algorithm domain.SigningAlgorithm,
) *EnsureSigningKeysInteractor {
	if logger == nil {
		logger = port.NopLogger{}
	}
	return &EnsureSigningKeysInteractor{
		signingKeyReader: signingKeyReader,
		signingKeyWriter: signingKeyWriter,
		keyPairGenerator: keyPairGenerator,
		clock:            clock,
		uuidGenerator:    uuidGenerator,
		logger:           logger,
		schedule:         schedule,
		algorithm:        algorithm,
	}
}

func (i *EnsureSigningKeysInteractor) Execute(ctx context.Context) error {
	now := i.clock.Now()
	latest, err := i.signingKeyReader.FindLatest(ctx)
	if err != nil {
		return err
	}
	window, reason, planned := i.schedule.PlanNext(latest, now)
	if !planned {
		return nil
	}
	keyPair, err := i.keyPairGenerator.Generate(ctx, i.algorithm)
	if err != nil {
		return domain.Internal(err)
	}
	key, err := i.signingKeyWriter.Create(ctx, port.SigningKeyCreate{
		ID:      i.uuidGenerator.NewUUID(),
		KeyPair: keyPair,
		Window:  window,
	})
	if err != nil {
		return err
	}
	fields := []any{
		"kid", key.KeyID,
		"reason", reason,
		"algorithm", key.Algorithm,
		"introduced_at", key.Window.IntroducedAt,
		"active_from", key.Window.ActiveFrom,
		"active_until", key.Window.ActiveUntil,
		"delete_after", key.Window.DeleteAfter,
	}
	switch reason {
	case domain.SigningKeyPlanReasonBootstrap:
		i.logger.Info(ctx, "first signing key created; generated a key pair, stored it, published it in the JWKS and it signs every new token from now", fields...)
	case domain.SigningKeyPlanReasonRecovery:
		i.logger.Warn(ctx, "signing key created without a pending phase because the latest key no longer signs; it signs every new token from now and peers must refresh on the unknown kid",
			append(fields, "replaces_kid", latest.KeyID)...)
	default:
		i.logger.Info(ctx, "next signing key created as pending; generated a key pair, stored it and published it in the JWKS so peers fetch it before it signs at active_from",
			append(fields, "replaces_kid", latest.KeyID)...)
	}
	return nil
}
