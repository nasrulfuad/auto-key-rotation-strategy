package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

var baseTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func testSchedule() domain.SigningKeySchedule {
	return domain.SigningKeySchedule{PendingDuration: 5 * time.Minute, ActiveDuration: 20 * time.Minute, InactiveDuration: 10 * time.Minute}
}

func keyWithWindow(window domain.SigningKeyWindow) domain.SigningKey {
	return domain.SigningKey{KeyID: "kid", Window: window}
}

func TestSigningKeyStatusAtBoundaries(t *testing.T) {
	key := keyWithWindow(domain.SigningKeyWindow{
		IntroducedAt: baseTime,
		ActiveFrom:   baseTime.Add(5 * time.Minute),
		ActiveUntil:  baseTime.Add(25 * time.Minute),
		DeleteAfter:  baseTime.Add(35 * time.Minute),
	})
	cases := []struct {
		name      string
		at        time.Time
		status    domain.SigningKeyStatus
		published bool
		canSign   bool
	}{
		{"introduced", baseTime, domain.SigningKeyStatusPending, true, false},
		{"just before active", baseTime.Add(5*time.Minute - time.Nanosecond), domain.SigningKeyStatusPending, true, false},
		{"exactly active_from", baseTime.Add(5 * time.Minute), domain.SigningKeyStatusActive, true, true},
		{"just before active_until", baseTime.Add(25*time.Minute - time.Nanosecond), domain.SigningKeyStatusActive, true, true},
		{"exactly active_until", baseTime.Add(25 * time.Minute), domain.SigningKeyStatusInactive, true, false},
		{"just before delete_after", baseTime.Add(35*time.Minute - time.Nanosecond), domain.SigningKeyStatusInactive, true, false},
		{"exactly delete_after", baseTime.Add(35 * time.Minute), domain.SigningKeyStatusExpired, false, false},
		{"before introduced", baseTime.Add(-time.Second), domain.SigningKeyStatusPending, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.status, key.StatusAt(tc.at))
			assert.Equal(t, tc.published, key.IsPublishedAt(tc.at))
			assert.Equal(t, tc.canSign, key.CanSignAt(tc.at))
		})
	}
}

func TestSigningKeySchedulePlanNextBootstrapsWhenNoKeyExists(t *testing.T) {
	window, reason, planned := testSchedule().PlanNext(nil, baseTime)

	require.True(t, planned)
	assert.Equal(t, domain.SigningKeyPlanReasonBootstrap, reason)
	assert.Equal(t, domain.SigningKeyWindow{
		IntroducedAt: baseTime,
		ActiveFrom:   baseTime,
		ActiveUntil:  baseTime.Add(20 * time.Minute),
		DeleteAfter:  baseTime.Add(30 * time.Minute),
	}, window)
}

func TestSigningKeySchedulePlanNextWaitsUntilPendingPhaseStarts(t *testing.T) {
	schedule := testSchedule()
	latest, _, _ := schedule.PlanNext(nil, baseTime)
	key := keyWithWindow(latest)

	_, _, planned := schedule.PlanNext(&key, baseTime.Add(15*time.Minute-time.Nanosecond))

	assert.False(t, planned)
}

func TestSigningKeySchedulePlanNextRotatesExactlyPendingDurationBeforeActiveUntil(t *testing.T) {
	schedule := testSchedule()
	latestWindow, _, _ := schedule.PlanNext(nil, baseTime)
	latest := keyWithWindow(latestWindow)
	now := baseTime.Add(15 * time.Minute)

	window, reason, planned := schedule.PlanNext(&latest, now)

	require.True(t, planned)
	assert.Equal(t, domain.SigningKeyPlanReasonRotation, reason)
	assert.Equal(t, now, window.IntroducedAt)
	assert.Equal(t, latest.Window.ActiveUntil, window.ActiveFrom)
	assert.Equal(t, latest.Window.ActiveUntil.Add(20*time.Minute), window.ActiveUntil)
	assert.Equal(t, window.ActiveUntil.Add(10*time.Minute), window.DeleteAfter)
}

func TestSigningKeySchedulePlanNextRecoversAfterDowntime(t *testing.T) {
	schedule := testSchedule()
	latestWindow, _, _ := schedule.PlanNext(nil, baseTime)
	latest := keyWithWindow(latestWindow)
	now := baseTime.Add(2 * time.Hour)

	window, reason, planned := schedule.PlanNext(&latest, now)

	require.True(t, planned)
	assert.Equal(t, domain.SigningKeyPlanReasonRecovery, reason)
	assert.Equal(t, now, window.ActiveFrom)
}

func TestSigningKeySchedulePlanNextKeepsWindowsContiguousOverManyRotations(t *testing.T) {
	schedule := testSchedule()
	var keys []domain.SigningKey
	for now := baseTime; now.Before(baseTime.Add(3 * time.Hour)); now = now.Add(15 * time.Second) {
		var latest *domain.SigningKey
		if len(keys) > 0 {
			latest = &keys[len(keys)-1]
		}
		if window, _, planned := schedule.PlanNext(latest, now); planned {
			keys = append(keys, keyWithWindow(window))
		}
		active := 0
		for _, key := range keys {
			if key.CanSignAt(now) {
				active++
			}
		}
		require.Equal(t, 1, active, "exactly one key must be able to sign at %s", now)
	}
	for i := 1; i < len(keys); i++ {
		assert.Equal(t, keys[i-1].Window.ActiveUntil, keys[i].Window.ActiveFrom)
		assert.Equal(t, 5*time.Minute, keys[i].Window.ActiveFrom.Sub(keys[i].Window.IntroducedAt))
	}
}

func TestSigningKeyScheduleValidateRejectsTokenOutlivingInactivePhase(t *testing.T) {
	schedule := testSchedule()

	assert.NoError(t, schedule.Validate(10*time.Minute))
	assert.Error(t, schedule.Validate(10*time.Minute+time.Second))
	assert.Error(t, schedule.Validate(0))
	assert.Error(t, domain.SigningKeySchedule{PendingDuration: 0, ActiveDuration: time.Minute, InactiveDuration: time.Minute}.Validate(time.Second))
	long := domain.SigningKeySchedule{PendingDuration: time.Hour, ActiveDuration: time.Hour, InactiveDuration: 48 * time.Hour}
	assert.Error(t, long.Validate(25*time.Hour))
}
