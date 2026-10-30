package domain

import (
	"time"

	"github.com/google/uuid"
)

type FreeTrialStatus string

const (
	FreeTrialStatusActive   FreeTrialStatus = "ACTIVE"
	FreeTrialStatusInactive FreeTrialStatus = "INACTIVE"
	FreeTrialStatusNone     FreeTrialStatus = "NONE"
	FreeTrialStatusUnknown  FreeTrialStatus = "UNKNOWN"
)

func (s FreeTrialStatus) Valid() bool {
	switch s {
	case FreeTrialStatusActive, FreeTrialStatusInactive, FreeTrialStatusNone, FreeTrialStatusUnknown:
		return true
	default:
		return false
	}
}

type EntitlementStatus string

const (
	EntitlementStatusActive     EntitlementStatus = "ACTIVE"
	EntitlementStatusTerminated EntitlementStatus = "TERMINATED"
)

const MaximumProductsLength = 255

type Entitlement struct {
	ID            uuid.UUID
	Provider      string
	SubjectID     string
	EntitlementID string
	Products      string
	FreeTrial     FreeTrialStatus
	Status        EntitlementStatus
	TerminatedAt  *time.Time
	LastUpdatedAt time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type EntitlementChange struct {
	Provider      string
	SubjectID     string
	EntitlementID string
	Products      string
	FreeTrial     FreeTrialStatus
	LastUpdatedAt time.Time
}

func NewEntitlementChange(provider, subjectID, entitlementID, products, freeTrial string, lastUpdatedAt time.Time) (EntitlementChange, error) {
	if err := ValidateSubjectID(subjectID); err != nil {
		return EntitlementChange{}, err
	}
	if err := ValidateEntitlementID(entitlementID); err != nil {
		return EntitlementChange{}, err
	}
	if products == "" {
		return EntitlementChange{}, Invalid("products", "products is required")
	}
	if len(products) > MaximumProductsLength {
		return EntitlementChange{}, Invalid("products", "products must be at most 255 characters")
	}
	status := FreeTrialStatus(freeTrial)
	if status == "" {
		status = FreeTrialStatusUnknown
	}
	if !status.Valid() {
		return EntitlementChange{}, Invalid("offer.free_trial", "offer.free_trial must be one of ACTIVE, INACTIVE, NONE, UNKNOWN")
	}
	return EntitlementChange{
		Provider:      provider,
		SubjectID:     subjectID,
		EntitlementID: entitlementID,
		Products:      products,
		FreeTrial:     status,
		LastUpdatedAt: lastUpdatedAt,
	}, nil
}

type EntitlementTermination struct {
	Provider      string
	SubjectID     string
	EntitlementID string
	LastUpdatedAt time.Time
}

func NewEntitlementTermination(provider, subjectID, entitlementID string, lastUpdatedAt time.Time) (EntitlementTermination, error) {
	if err := ValidateSubjectID(subjectID); err != nil {
		return EntitlementTermination{}, err
	}
	if err := ValidateEntitlementID(entitlementID); err != nil {
		return EntitlementTermination{}, err
	}
	return EntitlementTermination{Provider: provider, SubjectID: subjectID, EntitlementID: entitlementID, LastUpdatedAt: lastUpdatedAt}, nil
}

var (
	ErrEntitlementNotFound    = NotFoundCoded("ENTITLEMENT_NOT_FOUND", "entitlement_not_found")
	ErrEntitlementStaleUpdate = BadRequestCoded("ENTITLEMENT_STALE_UPDATE",
		"the token was issued before the last update of this entitlement")
)
