package domain

import (
	"time"

	"github.com/google/uuid"
)

type ActivationNotificationStatus string

const (
	ActivationNotificationStatusPending ActivationNotificationStatus = "pending"
	ActivationNotificationStatusSent    ActivationNotificationStatus = "sent"
	ActivationNotificationStatusFailed  ActivationNotificationStatus = "failed"
)

type ActivationNotificationEvent string

const (
	ActivationNotificationEventEntitlementUpserted   ActivationNotificationEvent = "entitlement_upserted"
	ActivationNotificationEventEntitlementTerminated ActivationNotificationEvent = "entitlement_terminated"
)

const (
	activationNotificationBaseRetryDelay    = 5 * time.Second
	activationNotificationMaximumRetryDelay = 5 * time.Minute
)

type ActivationNotification struct {
	ID            uuid.UUID
	Provider      string
	SubjectID     string
	EntitlementID string
	Products      string
	FreeTrial     FreeTrialStatus
	Event         ActivationNotificationEvent
	Status        ActivationNotificationStatus
	Attempts      int
	NextAttemptAt time.Time
	CreatedAt     time.Time
}

type ActivationNotificationAttemptResult struct {
	Attempts      int
	Status        ActivationNotificationStatus
	NextAttemptAt time.Time
}

func (n ActivationNotification) FailAttempt(now time.Time, maximumAttempts int) ActivationNotificationAttemptResult {
	attempts := n.Attempts + 1
	if attempts >= maximumAttempts {
		return ActivationNotificationAttemptResult{Attempts: attempts, Status: ActivationNotificationStatusFailed, NextAttemptAt: now}
	}
	return ActivationNotificationAttemptResult{
		Attempts:      attempts,
		Status:        ActivationNotificationStatusPending,
		NextAttemptAt: now.Add(ActivationNotificationRetryDelay(attempts)),
	}
}

func ActivationNotificationRetryDelay(attempts int) time.Duration {
	delay := activationNotificationBaseRetryDelay
	for range attempts - 1 {
		delay *= 2
		if delay >= activationNotificationMaximumRetryDelay {
			return activationNotificationMaximumRetryDelay
		}
	}
	return delay
}
