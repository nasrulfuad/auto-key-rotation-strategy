package domain

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID              uuid.UUID
	SubjectID       string
	EntitlementID   string
	Products        string
	FreeTrial       FreeTrialStatus
	Status          EntitlementStatus
	SignedWithKeyID string
	ActivatedAt     *time.Time
	TerminatedAt    *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewSubscriptionFromPartner(id uuid.UUID, entitlement PartnerEntitlement, signedWithKeyID string, now time.Time) (Subscription, error) {
	if !entitlement.Status.Valid() {
		return Subscription{}, Internalf("partner_app returned an unknown entitlement status %q", entitlement.Status)
	}
	subscription := Subscription{
		ID:              id,
		SubjectID:       entitlement.SubjectID,
		EntitlementID:   entitlement.EntitlementID,
		Products:        entitlement.Products,
		FreeTrial:       entitlement.FreeTrial,
		Status:          entitlement.Status,
		SignedWithKeyID: signedWithKeyID,
	}
	switch entitlement.Status {
	case EntitlementStatusActive:
		subscription.ActivatedAt = &now
	case EntitlementStatusTerminated:
		terminatedAt := now
		if entitlement.TerminatedAt != nil {
			terminatedAt = *entitlement.TerminatedAt
		}
		subscription.TerminatedAt = &terminatedAt
	}
	return subscription, nil
}
