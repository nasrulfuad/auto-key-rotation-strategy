package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
)

func TestActivationNotificationRetryDelayBacksOffExponentially(t *testing.T) {
	assert.Equal(t, 5*time.Second, domain.ActivationNotificationRetryDelay(1))
	assert.Equal(t, 10*time.Second, domain.ActivationNotificationRetryDelay(2))
	assert.Equal(t, 20*time.Second, domain.ActivationNotificationRetryDelay(3))
	assert.Equal(t, 5*time.Minute, domain.ActivationNotificationRetryDelay(20))
}

func TestActivationNotificationFailAttemptMarksFailedAtMaximum(t *testing.T) {
	notification := domain.ActivationNotification{Attempts: 3}

	retry := notification.FailAttempt(baseTime, 5)
	assert.Equal(t, domain.ActivationNotificationAttemptResult{
		Attempts: 4, Status: domain.ActivationNotificationStatusPending, NextAttemptAt: baseTime.Add(40 * time.Second),
	}, retry)

	notification.Attempts = 4
	failed := notification.FailAttempt(baseTime, 5)
	assert.Equal(t, domain.ActivationNotificationStatusFailed, failed.Status)
	assert.Equal(t, 5, failed.Attempts)
}
