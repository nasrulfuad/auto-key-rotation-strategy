package domain

import (
	"time"

	"github.com/google/uuid"
)

type ActivationNotificationEvent string

const (
	ActivationNotificationEventEntitlementUpserted   ActivationNotificationEvent = "entitlement_upserted"
	ActivationNotificationEventEntitlementTerminated ActivationNotificationEvent = "entitlement_terminated"
)

func (e ActivationNotificationEvent) Valid() bool {
	return e == ActivationNotificationEventEntitlementUpserted || e == ActivationNotificationEventEntitlementTerminated
}

type ActivationNotification struct {
	ID             uuid.UUID
	NotificationID uuid.UUID
	KeyID          string
	Issuer         string
	Event          ActivationNotificationEvent
	Provider       string
	SubjectID      string
	EntitlementID  string
	Products       string
	FreeTrial      FreeTrialStatus
	ReceivedAt     time.Time
}

func (n ActivationNotification) Validate(token VerifiedToken) error {
	if n.NotificationID == uuid.Nil {
		return Invalid("notification_id", "notification_id must be a UUID")
	}
	if !n.Event.Valid() {
		return Invalid("event", "event is not supported")
	}
	if err := ValidateSubjectID(n.SubjectID); err != nil {
		return err
	}
	if err := ValidateEntitlementID(n.EntitlementID); err != nil {
		return err
	}
	if token.Claims.Subject != n.SubjectID {
		return TokenClaimInvalid("subject_mismatch")
	}
	if token.Claims.Provider != n.Provider {
		return TokenClaimInvalid("provider_mismatch")
	}
	return nil
}
