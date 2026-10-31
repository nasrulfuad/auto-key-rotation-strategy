package signingkey_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging/loggingmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/signingkey"
)

type trackFixture struct {
	reader *repositorymocks.MockSigningKeyReader
	clock  *systemclockmocks.MockClock
	logger *loggingmocks.MockLogger
	uc     *signingkey.TrackSigningKeyStatusesInteractor
}

func newTrackFixture(t *testing.T) trackFixture {
	f := trackFixture{
		reader: repositorymocks.NewMockSigningKeyReader(t),
		clock:  systemclockmocks.NewMockClock(t),
		logger: loggingmocks.NewMockLogger(t),
	}
	f.uc = signingkey.NewTrackSigningKeyStatusesInteractor(f.reader, signingkey.WithClock(f.clock), signingkey.WithLogger(f.logger))
	return f
}

func signingKeyWindowFrom(activeFrom time.Time) domain.SigningKeyWindow {
	return domain.SigningKeyWindow{
		IntroducedAt: activeFrom.Add(-5 * time.Minute),
		ActiveFrom:   activeFrom,
		ActiveUntil:  activeFrom.Add(20 * time.Minute),
		DeleteAfter:  activeFrom.Add(30 * time.Minute),
	}
}

func expectStatusLog(f trackFixture, message string, key domain.SigningKey, status domain.SigningKeyStatus, extra ...any) {
	fields := []any{
		"kid", key.KeyID, "status", status,
		"active_from", key.Window.ActiveFrom, "active_until", key.Window.ActiveUntil, "delete_after", key.Window.DeleteAfter}
	f.logger.EXPECT().Info(mock.Anything, message, append(fields, extra...)).Return().Once()
}

func TestTrackSigningKeyStatusesLogsEveryKeyOnFirstObservation(t *testing.T) {
	f := newTrackFixture(t)
	current := domain.SigningKey{KeyID: "current", Window: signingKeyWindowFrom(now.Add(-time.Minute))}
	next := domain.SigningKey{KeyID: "next", Window: signingKeyWindowFrom(now.Add(time.Minute))}
	f.clock.EXPECT().Now().Return(now).Once()
	f.reader.EXPECT().ListAll(mock.Anything).Return([]domain.SigningKey{current, next}, nil).Once()
	expectStatusLog(f, "signing key status observed at startup", current, domain.SigningKeyStatusActive)
	expectStatusLog(f, "signing key status observed at startup", next, domain.SigningKeyStatusPending)

	require.NoError(t, f.uc.Execute(context.Background()))
}

func TestTrackSigningKeyStatusesLogsActivationAndRetirement(t *testing.T) {
	f := newTrackFixture(t)
	current := domain.SigningKey{KeyID: "current", Window: signingKeyWindowFrom(now.Add(-19 * time.Minute))}
	next := domain.SigningKey{KeyID: "next", Window: signingKeyWindowFrom(current.Window.ActiveUntil)}
	keys := []domain.SigningKey{current, next}
	f.reader.EXPECT().ListAll(mock.Anything).Return(keys, nil).Times(2)
	f.clock.EXPECT().Now().Return(now).Once()
	expectStatusLog(f, "signing key status observed at startup", current, domain.SigningKeyStatusActive)
	expectStatusLog(f, "signing key status observed at startup", next, domain.SigningKeyStatusPending)
	require.NoError(t, f.uc.Execute(context.Background()))

	f.clock.EXPECT().Now().Return(now.Add(2 * time.Minute)).Once()
	expectStatusLog(f, "signing key retired from signing; it stays in the JWKS only to verify tokens it already signed until delete_after",
		current, domain.SigningKeyStatusInactive, "previous_status", domain.SigningKeyStatusActive)
	expectStatusLog(f, "signing key activated; it now signs every new token",
		next, domain.SigningKeyStatusActive, "previous_status", domain.SigningKeyStatusPending)
	require.NoError(t, f.uc.Execute(context.Background()))
}

func TestTrackSigningKeyStatusesSkipsUnchangedAndNewPendingKeys(t *testing.T) {
	f := newTrackFixture(t)
	current := domain.SigningKey{KeyID: "current", Window: signingKeyWindowFrom(now.Add(-time.Minute))}
	created := domain.SigningKey{KeyID: "created", Window: signingKeyWindowFrom(now.Add(5 * time.Minute))}
	f.clock.EXPECT().Now().Return(now).Times(2)
	f.reader.EXPECT().ListAll(mock.Anything).Return([]domain.SigningKey{current}, nil).Once()
	expectStatusLog(f, "signing key status observed at startup", current, domain.SigningKeyStatusActive)
	require.NoError(t, f.uc.Execute(context.Background()))

	f.reader.EXPECT().ListAll(mock.Anything).Return([]domain.SigningKey{current, created}, nil).Once()
	require.NoError(t, f.uc.Execute(context.Background()))
}

func TestTrackSigningKeyStatusesLogsExpiry(t *testing.T) {
	f := newTrackFixture(t)
	retired := domain.SigningKey{KeyID: "retired", Window: signingKeyWindowFrom(now.Add(-29 * time.Minute))}
	f.reader.EXPECT().ListAll(mock.Anything).Return([]domain.SigningKey{retired}, nil).Times(2)
	f.clock.EXPECT().Now().Return(now).Once()
	expectStatusLog(f, "signing key status observed at startup", retired, domain.SigningKeyStatusInactive)
	require.NoError(t, f.uc.Execute(context.Background()))

	f.clock.EXPECT().Now().Return(now.Add(2 * time.Minute)).Once()
	expectStatusLog(f, "signing key expired; it is no longer published in the JWKS and the delete job removes it on its next tick",
		retired, domain.SigningKeyStatusExpired, "previous_status", domain.SigningKeyStatusInactive)
	require.NoError(t, f.uc.Execute(context.Background()))
}
