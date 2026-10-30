package domain

import "time"

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

func (s EntitlementStatus) Valid() bool {
	return s == EntitlementStatusActive || s == EntitlementStatusTerminated
}

type PartnerEntitlement struct {
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

type PartnerEntitlementChange struct {
	SubjectID     string
	EntitlementID string
	Products      string
	FreeTrial     FreeTrialStatus
}

func NewPartnerEntitlementChange(subjectID, entitlementID, products, freeTrial string) (PartnerEntitlementChange, error) {
	if err := ValidateSubjectID(subjectID); err != nil {
		return PartnerEntitlementChange{}, err
	}
	if err := ValidateEntitlementID(entitlementID); err != nil {
		return PartnerEntitlementChange{}, err
	}
	if products == "" {
		return PartnerEntitlementChange{}, Invalid("products", "products is required")
	}
	status := FreeTrialStatus(freeTrial)
	if status != "" && !status.Valid() {
		return PartnerEntitlementChange{}, Invalid("offer.free_trial", "offer.free_trial must be one of ACTIVE, INACTIVE, NONE, UNKNOWN")
	}
	return PartnerEntitlementChange{SubjectID: subjectID, EntitlementID: entitlementID, Products: products, FreeTrial: status}, nil
}

func ValidateEntitlementReference(subjectID, entitlementID string) error {
	if err := ValidateSubjectID(subjectID); err != nil {
		return err
	}
	return ValidateEntitlementID(entitlementID)
}
