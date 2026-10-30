package signingkey_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging/loggingmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/rsakey/rsakeymocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/signingkey"
)

const rotationLockName = "ensure_signing_keys"

type rotateFixture struct {
	jobLock   *repositorymocks.MockJobLock
	reader    *repositorymocks.MockSigningKeyReader
	writer    *repositorymocks.MockSigningKeyWriter
	generator *rsakeymocks.MockKeyPairGenerator
	clock     *systemclockmocks.MockClock
	uuids     *systemclockmocks.MockUUIDGenerator
	logger    *loggingmocks.MockLogger
	uc        *signingkey.RotateSigningKeyInteractor
}

func newRotateFixture(t *testing.T) rotateFixture {
	f := rotateFixture{
		jobLock:   repositorymocks.NewMockJobLock(t),
		reader:    repositorymocks.NewMockSigningKeyReader(t),
		writer:    repositorymocks.NewMockSigningKeyWriter(t),
		generator: rsakeymocks.NewMockKeyPairGenerator(t),
		clock:     systemclockmocks.NewMockClock(t),
		uuids:     systemclockmocks.NewMockUUIDGenerator(t),
		logger:    loggingmocks.NewMockLogger(t),
	}
	f.uc = signingkey.NewRotateSigningKeyInteractor(f.jobLock, f.reader, f.writer, f.generator, f.clock, f.uuids, f.logger,
		rotationLockName, schedule, domain.SigningAlgorithmRS256)
	return f
}

func (f rotateFixture) expectLockAcquired() {
	f.jobLock.EXPECT().WithLock(mock.Anything, rotationLockName, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, run func(ctx context.Context) error) (bool, error) {
			return true, run(ctx)
		})
	f.clock.EXPECT().Now().Return(now)
}

func activeKeyAt(moment time.Time) domain.SigningKey {
	return domain.SigningKey{ID: uuid.New(), KeyID: "current", Window: domain.SigningKeyWindow{
		IntroducedAt: moment.Add(-10 * time.Minute), ActiveFrom: moment.Add(-10 * time.Minute),
		ActiveUntil: moment.Add(10 * time.Minute), DeleteAfter: moment.Add(20 * time.Minute),
	}}
}

func TestRotateSigningKeyCreatesPendingKeyAndShortensCurrentKey(t *testing.T) {
	f := newRotateFixture(t)
	f.expectLockAcquired()
	current := activeKeyAt(now)
	keyPair := domain.KeyPair{KeyID: "next", Algorithm: domain.SigningAlgorithmRS256}
	id := uuid.New()
	activeFrom := now.Add(5 * time.Minute)
	nextWindow := domain.SigningKeyWindow{
		IntroducedAt: now, ActiveFrom: activeFrom,
		ActiveUntil: activeFrom.Add(20 * time.Minute), DeleteAfter: activeFrom.Add(30 * time.Minute),
	}
	currentWindow := current.Window
	currentWindow.ActiveUntil = activeFrom
	currentWindow.DeleteAfter = activeFrom.Add(10 * time.Minute)
	f.reader.EXPECT().FindLatest(mock.Anything).Return(&current, nil)
	f.generator.EXPECT().Generate(mock.Anything, domain.SigningAlgorithmRS256).Return(keyPair, nil)
	f.uuids.EXPECT().NewUUID().Return(id)
	nextKey := domain.SigningKey{ID: id, KeyID: "next", Algorithm: domain.SigningAlgorithmRS256, Window: nextWindow}
	f.writer.EXPECT().Rotate(mock.Anything, repository.SigningKeyRotate{
		CurrentKeyID:     current.ID,
		CurrentKeyWindow: currentWindow,
		NextKey:          repository.SigningKeyCreate{ID: id, KeyPair: keyPair, Window: nextWindow},
	}).Return(nextKey, nil)
	f.logger.EXPECT().Info(mock.Anything,
		"signing key rotation requested; next signing key created as pending and the current key stops signing when it becomes active",
		[]any{
			"kid", "next",
			"reason", "manual",
			"algorithm", domain.SigningAlgorithmRS256,
			"introduced_at", nextWindow.IntroducedAt,
			"active_from", nextWindow.ActiveFrom,
			"active_until", nextWindow.ActiveUntil,
			"delete_after", nextWindow.DeleteAfter,
			"replaces_kid", "current",
			"replaced_active_until", currentWindow.ActiveUntil,
			"replaced_delete_after", currentWindow.DeleteAfter,
		}).Return()

	out, err := f.uc.Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, now, out.Now)
	assert.Equal(t, nextKey, out.NextKey)
	assert.Equal(t, "current", out.CurrentKey.KeyID)
	assert.Equal(t, currentWindow, out.CurrentKey.Window)
	assert.Equal(t, domain.SigningKeyStatusPending, out.NextKey.StatusAt(now))
}

func TestRotateSigningKeyRejectsWhenPendingKeyExists(t *testing.T) {
	f := newRotateFixture(t)
	f.expectLockAcquired()
	pending := domain.SigningKey{KeyID: "pending", Window: domain.SigningKeyWindow{
		IntroducedAt: now.Add(-time.Minute), ActiveFrom: now.Add(4 * time.Minute),
		ActiveUntil: now.Add(24 * time.Minute), DeleteAfter: now.Add(34 * time.Minute),
	}}
	f.reader.EXPECT().FindLatest(mock.Anything).Return(&pending, nil)

	_, err := f.uc.Execute(context.Background())

	require.ErrorIs(t, err, domain.ErrSigningKeyAlreadyPending)
}

func TestRotateSigningKeyRejectsWithoutActiveKey(t *testing.T) {
	f := newRotateFixture(t)
	f.expectLockAcquired()
	f.reader.EXPECT().FindLatest(mock.Anything).Return(nil, nil)

	_, err := f.uc.Execute(context.Background())

	require.ErrorIs(t, err, domain.ErrSigningKeyNotFound)
}

func TestRotateSigningKeyRejectsWhileLockIsHeld(t *testing.T) {
	f := newRotateFixture(t)
	f.jobLock.EXPECT().WithLock(mock.Anything, rotationLockName, mock.Anything).Return(false, nil)

	_, err := f.uc.Execute(context.Background())

	require.ErrorIs(t, err, domain.ErrSigningKeyRotationInProgress)
}

func TestRotateSigningKeyReturnsWriterError(t *testing.T) {
	f := newRotateFixture(t)
	f.expectLockAcquired()
	current := activeKeyAt(now)
	writeErr := domain.DB(errors.New("connection reset"))
	f.reader.EXPECT().FindLatest(mock.Anything).Return(&current, nil)
	f.generator.EXPECT().Generate(mock.Anything, domain.SigningAlgorithmRS256).Return(domain.KeyPair{KeyID: "next"}, nil)
	f.uuids.EXPECT().NewUUID().Return(uuid.New())
	f.writer.EXPECT().Rotate(mock.Anything, mock.Anything).Return(domain.SigningKey{}, writeErr)

	_, err := f.uc.Execute(context.Background())

	require.ErrorIs(t, err, writeErr)
}
