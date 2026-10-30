package signingkey_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/jwt/jwtmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging/loggingmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/rsakey/rsakeymocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/signingkey"
)

var now = time.Date(2026, 1, 1, 0, 15, 0, 0, time.UTC)

var schedule = domain.SigningKeySchedule{PendingDuration: 5 * time.Minute, ActiveDuration: 20 * time.Minute, InactiveDuration: 10 * time.Minute}

type ensureFixture struct {
	reader    *repositorymocks.MockSigningKeyReader
	writer    *repositorymocks.MockSigningKeyWriter
	generator *rsakeymocks.MockKeyPairGenerator
	clock     *systemclockmocks.MockClock
	uuids     *systemclockmocks.MockUUIDGenerator
	logger    *loggingmocks.MockLogger
	uc        *signingkey.EnsureSigningKeysInteractor
}

func newEnsureFixture(t *testing.T) ensureFixture {
	f := ensureFixture{
		reader:    repositorymocks.NewMockSigningKeyReader(t),
		writer:    repositorymocks.NewMockSigningKeyWriter(t),
		generator: rsakeymocks.NewMockKeyPairGenerator(t),
		clock:     systemclockmocks.NewMockClock(t),
		uuids:     systemclockmocks.NewMockUUIDGenerator(t),
		logger:    loggingmocks.NewMockLogger(t),
	}
	f.uc = signingkey.NewEnsureSigningKeysInteractor(f.reader, f.writer, f.generator, f.clock, f.uuids, f.logger, schedule, domain.SigningAlgorithmRS256)
	f.clock.EXPECT().Now().Return(now)
	return f
}

func createdKeyFields(keyID string, reason domain.SigningKeyPlanReason, window domain.SigningKeyWindow, extra ...any) []any {
	fields := []any{
		"kid", keyID,
		"reason", reason,
		"algorithm", domain.SigningAlgorithm(""),
		"introduced_at", window.IntroducedAt,
		"active_from", window.ActiveFrom,
		"active_until", window.ActiveUntil,
		"delete_after", window.DeleteAfter,
	}
	return append(fields, extra...)
}

func TestEnsureSigningKeysCreatesNextKeyWhenPendingPhaseStarts(t *testing.T) {
	f := newEnsureFixture(t)
	latest := domain.SigningKey{KeyID: "old", Window: domain.SigningKeyWindow{
		IntroducedAt: now.Add(-15 * time.Minute), ActiveFrom: now.Add(-15 * time.Minute),
		ActiveUntil: now.Add(5 * time.Minute), DeleteAfter: now.Add(15 * time.Minute),
	}}
	keyPair := domain.KeyPair{KeyID: "new", Algorithm: domain.SigningAlgorithmRS256}
	id := uuid.New()
	f.reader.EXPECT().FindLatest(mock.Anything).Return(&latest, nil)
	f.generator.EXPECT().Generate(mock.Anything, domain.SigningAlgorithmRS256).Return(keyPair, nil)
	f.uuids.EXPECT().NewUUID().Return(id)
	wantWindow := domain.SigningKeyWindow{
		IntroducedAt: now, ActiveFrom: latest.Window.ActiveUntil,
		ActiveUntil: latest.Window.ActiveUntil.Add(20 * time.Minute), DeleteAfter: latest.Window.ActiveUntil.Add(30 * time.Minute),
	}
	f.writer.EXPECT().Create(mock.Anything, repository.SigningKeyCreate{ID: id, KeyPair: keyPair, Window: wantWindow}).
		Return(domain.SigningKey{KeyID: "new", Window: wantWindow}, nil)
	f.logger.EXPECT().Info(mock.Anything,
		"next signing key created as pending; generated a key pair, stored it and published it in the JWKS so peers fetch it before it signs at active_from",
		createdKeyFields("new", domain.SigningKeyPlanReasonRotation, wantWindow, "replaces_kid", "old")).Return()

	require.NoError(t, f.uc.Execute(context.Background()))
}

func TestEnsureSigningKeysDoesNothingBeforePendingPhase(t *testing.T) {
	f := newEnsureFixture(t)
	latest := domain.SigningKey{KeyID: "current", Window: domain.SigningKeyWindow{
		IntroducedAt: now, ActiveFrom: now, ActiveUntil: now.Add(20 * time.Minute), DeleteAfter: now.Add(30 * time.Minute),
	}}
	f.reader.EXPECT().FindLatest(mock.Anything).Return(&latest, nil)

	require.NoError(t, f.uc.Execute(context.Background()))
}

func TestEnsureSigningKeysWarnsOnRecovery(t *testing.T) {
	f := newEnsureFixture(t)
	latest := domain.SigningKey{KeyID: "stale", Window: domain.SigningKeyWindow{
		IntroducedAt: now.Add(-3 * time.Hour), ActiveFrom: now.Add(-3 * time.Hour),
		ActiveUntil: now.Add(-2 * time.Hour), DeleteAfter: now.Add(-time.Hour),
	}}
	f.reader.EXPECT().FindLatest(mock.Anything).Return(&latest, nil)
	f.generator.EXPECT().Generate(mock.Anything, domain.SigningAlgorithmRS256).Return(domain.KeyPair{KeyID: "fresh"}, nil)
	f.uuids.EXPECT().NewUUID().Return(uuid.New())
	f.writer.EXPECT().Create(mock.Anything, mock.MatchedBy(func(in repository.SigningKeyCreate) bool {
		return in.Window.ActiveFrom.Equal(now) && in.Window.IntroducedAt.Equal(now)
	})).Return(domain.SigningKey{KeyID: "fresh"}, nil)
	f.logger.EXPECT().Warn(mock.Anything,
		"signing key created without a pending phase because the latest key no longer signs; it signs every new token from now and peers must refresh on the unknown kid",
		createdKeyFields("fresh", domain.SigningKeyPlanReasonRecovery, domain.SigningKeyWindow{}, "replaces_kid", "stale")).Return()

	require.NoError(t, f.uc.Execute(context.Background()))
}

func TestEnsureSigningKeysLogsBootstrapWhenNoKeyExists(t *testing.T) {
	f := newEnsureFixture(t)
	f.reader.EXPECT().FindLatest(mock.Anything).Return(nil, nil)
	f.generator.EXPECT().Generate(mock.Anything, domain.SigningAlgorithmRS256).Return(domain.KeyPair{KeyID: "first"}, nil)
	f.uuids.EXPECT().NewUUID().Return(uuid.New())
	f.writer.EXPECT().Create(mock.Anything, mock.Anything).Return(domain.SigningKey{KeyID: "first"}, nil)
	f.logger.EXPECT().Info(mock.Anything,
		"first signing key created; generated a key pair, stored it, published it in the JWKS and it signs every new token from now",
		createdKeyFields("first", domain.SigningKeyPlanReasonBootstrap, domain.SigningKeyWindow{})).Return()

	require.NoError(t, f.uc.Execute(context.Background()))
}

func TestDeleteExpiredSigningKeysLogsEveryDeletedKey(t *testing.T) {
	writer := repositorymocks.NewMockSigningKeyWriter(t)
	clock := systemclockmocks.NewMockClock(t)
	logger := loggingmocks.NewMockLogger(t)
	clock.EXPECT().Now().Return(now)
	writer.EXPECT().DeleteExpired(mock.Anything, now).Return([]string{"a", "b"}, nil)
	for _, keyID := range []string{"a", "b"} {
		logger.EXPECT().Info(mock.Anything,
			"expired signing key deleted; removed from the database and the JWKS, tokens signed by it are now rejected",
			[]any{"kid", keyID, "deleted_at", now}).Return().Once()
	}

	err := signingkey.NewDeleteExpiredSigningKeysInteractor(writer, clock, logger).Execute(context.Background())

	assert.NoError(t, err)
}

func TestSignerUsesActiveKeyAndTruncatesIssuedAt(t *testing.T) {
	reader := repositorymocks.NewMockSigningKeyReader(t)
	issuer := jwtmocks.NewMockTokenIssuer(t)
	clock := systemclockmocks.NewMockClock(t)
	uuids := systemclockmocks.NewMockUUIDGenerator(t)
	instant := now.Add(1500 * time.Millisecond)
	jti := uuid.New()
	key := domain.SigningKey{KeyID: "active"}
	clock.EXPECT().Now().Return(instant)
	uuids.EXPECT().NewUUID().Return(jti)
	reader.EXPECT().FindActive(mock.Anything, instant).Return(key, nil)
	wantClaims := domain.TokenClaims{
		ID: jti.String(), Issuer: "qa.partner-app.local", Subject: "SUB1", Provider: "my_app",
		Audience: []string{"aud"}, IssuedAt: now.Add(time.Second), ExpiresAt: now.Add(time.Second + 5*time.Minute),
	}
	issuer.EXPECT().Issue(mock.Anything, wantClaims, key).Return("token", nil)

	signed, err := signingkey.NewSigner(reader, issuer, clock, uuids, "qa.partner-app.local", 5*time.Minute).
		Sign(context.Background(), domain.TokenRequest{Audience: "aud", Subject: "SUB1", Provider: "my_app"})

	require.NoError(t, err)
	assert.Equal(t, "token", signed.Value)
	assert.Equal(t, "active", signed.KeyID)
}
