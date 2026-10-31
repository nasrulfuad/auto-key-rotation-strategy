package signingkey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/systemclock"
)

type TrackSigningKeyStatusesInteractor struct {
	signingKeyReader repository.SigningKeyReader
	clock            systemclock.Clock
	logger           logging.Logger

	observedStatuses map[string]domain.SigningKeyStatus
}

func NewTrackSigningKeyStatusesInteractor(signingKeyReader repository.SigningKeyReader, opts ...Option) *TrackSigningKeyStatusesInteractor {
	o := newOptions(opts)
	return &TrackSigningKeyStatusesInteractor{signingKeyReader: signingKeyReader, clock: o.clock, logger: o.logger}
}

func (i *TrackSigningKeyStatusesInteractor) Execute(ctx context.Context) error {
	now := i.clock.Now()
	keys, err := i.signingKeyReader.ListAll(ctx)
	if err != nil {
		return err
	}
	firstObservation := i.observedStatuses == nil
	statuses := make(map[string]domain.SigningKeyStatus, len(keys))
	for _, key := range keys {
		status := key.StatusAt(now)
		statuses[key.KeyID] = status
		previous, observed := i.observedStatuses[key.KeyID]
		switch {
		case firstObservation:
			i.logger.Info(ctx, "signing key status observed at startup", signingKeyStatusFields(key, status, "")...)
		case status == previous:
		case observed || status != domain.SigningKeyStatusPending:
			i.logStatusChange(ctx, key, status, previous)
		}
	}
	i.observedStatuses = statuses
	return nil
}

func (i *TrackSigningKeyStatusesInteractor) logStatusChange(ctx context.Context, key domain.SigningKey, status, previous domain.SigningKeyStatus) {
	fields := signingKeyStatusFields(key, status, previous)
	switch status {
	case domain.SigningKeyStatusPending:
		i.logger.Info(ctx, "signing key is pending; it is published in the JWKS but does not sign yet", fields...)
	case domain.SigningKeyStatusActive:
		i.logger.Info(ctx, "signing key activated; it now signs every new token", fields...)
	case domain.SigningKeyStatusInactive:
		i.logger.Info(ctx, "signing key retired from signing; it stays in the JWKS only to verify tokens it already signed until delete_after", fields...)
	case domain.SigningKeyStatusExpired:
		i.logger.Info(ctx, "signing key expired; it is no longer published in the JWKS and the delete job removes it on its next tick", fields...)
	}
}

func signingKeyStatusFields(key domain.SigningKey, status, previous domain.SigningKeyStatus) []any {
	fields := []any{
		"kid", key.KeyID,
		"status", status,
		"active_from", key.Window.ActiveFrom,
		"active_until", key.Window.ActiveUntil,
		"delete_after", key.Window.DeleteAfter,
	}
	if previous != "" {
		fields = append(fields, "previous_status", previous)
	}
	return fields
}
