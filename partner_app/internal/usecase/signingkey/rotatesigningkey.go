package signingkey

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type RotateSigningKeyOutput struct {
	Now        time.Time
	CurrentKey domain.SigningKey
	NextKey    domain.SigningKey
}

type RotateSigningKeyInteractor struct {
	jobLock          port.JobLock
	signingKeyReader port.SigningKeyReader
	signingKeyWriter port.SigningKeyWriter
	keyPairGenerator port.KeyPairGenerator
	clock            port.Clock
	uuidGenerator    port.UUIDGenerator
	logger           port.Logger

	lockName  string
	schedule  domain.SigningKeySchedule
	algorithm domain.SigningAlgorithm
}

func NewRotateSigningKeyInteractor(
	jobLock port.JobLock,
	signingKeyReader port.SigningKeyReader,
	signingKeyWriter port.SigningKeyWriter,
	keyPairGenerator port.KeyPairGenerator,
	clock port.Clock,
	uuidGenerator port.UUIDGenerator,
	logger port.Logger,
	lockName string,
	schedule domain.SigningKeySchedule,
	algorithm domain.SigningAlgorithm,
) *RotateSigningKeyInteractor {
	if logger == nil {
		logger = port.NopLogger{}
	}
	return &RotateSigningKeyInteractor{
		jobLock:          jobLock,
		signingKeyReader: signingKeyReader,
		signingKeyWriter: signingKeyWriter,
		keyPairGenerator: keyPairGenerator,
		clock:            clock,
		uuidGenerator:    uuidGenerator,
		logger:           logger,
		lockName:         lockName,
		schedule:         schedule,
		algorithm:        algorithm,
	}
}

func (i *RotateSigningKeyInteractor) Execute(ctx context.Context) (RotateSigningKeyOutput, error) {
	var out RotateSigningKeyOutput
	acquired, err := i.jobLock.WithLock(ctx, i.lockName, func(ctx context.Context) error {
		var err error
		out, err = i.rotate(ctx)
		return err
	})
	if err != nil {
		return RotateSigningKeyOutput{}, err
	}
	if !acquired {
		return RotateSigningKeyOutput{}, domain.ErrSigningKeyRotationInProgress
	}
	return out, nil
}

func (i *RotateSigningKeyInteractor) rotate(ctx context.Context) (RotateSigningKeyOutput, error) {
	now := i.clock.Now()
	latest, err := i.signingKeyReader.FindLatest(ctx)
	if err != nil {
		return RotateSigningKeyOutput{}, err
	}
	rotation, err := i.schedule.PlanRotation(latest, now)
	if err != nil {
		return RotateSigningKeyOutput{}, err
	}
	keyPair, err := i.keyPairGenerator.Generate(ctx, i.algorithm)
	if err != nil {
		return RotateSigningKeyOutput{}, domain.Internal(err)
	}
	nextKey, err := i.signingKeyWriter.Rotate(ctx, port.SigningKeyRotate{
		CurrentKeyID:     latest.ID,
		CurrentKeyWindow: rotation.CurrentKeyWindow,
		NextKey: port.SigningKeyCreate{
			ID:      i.uuidGenerator.NewUUID(),
			KeyPair: keyPair,
			Window:  rotation.NextKeyWindow,
		},
	})
	if err != nil {
		return RotateSigningKeyOutput{}, err
	}
	currentKey := *latest
	currentKey.Window = rotation.CurrentKeyWindow
	i.logger.Info(ctx, "signing key rotation requested; next signing key created as pending and the current key stops signing when it becomes active",
		"kid", nextKey.KeyID,
		"reason", "manual",
		"algorithm", nextKey.Algorithm,
		"introduced_at", nextKey.Window.IntroducedAt,
		"active_from", nextKey.Window.ActiveFrom,
		"active_until", nextKey.Window.ActiveUntil,
		"delete_after", nextKey.Window.DeleteAfter,
		"replaces_kid", currentKey.KeyID,
		"replaced_active_until", currentKey.Window.ActiveUntil,
		"replaced_delete_after", currentKey.Window.DeleteAfter,
	)
	return RotateSigningKeyOutput{Now: now, CurrentKey: currentKey, NextKey: nextKey}, nil
}
