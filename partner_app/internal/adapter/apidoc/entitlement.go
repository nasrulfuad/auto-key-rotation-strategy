package apidoc

import "time"

type Offer struct {
	FreeTrial string `json:"free_trial" example:"NONE" enums:"ACTIVE,INACTIVE,NONE,UNKNOWN"`
}

type Entitlement struct {
	Provider      string     `json:"provider" example:"my_app"`
	SubjectID     string     `json:"subject_id" example:"SUB001"`
	EntitlementID string     `json:"entitlement_id" example:"ENT001"`
	Products      string     `json:"products" example:"com.partner.product1"`
	Offer         Offer      `json:"offer"`
	Status        string     `json:"status" example:"ACTIVE" enums:"ACTIVE,TERMINATED"`
	TerminatedAt  *time.Time `json:"terminated_at"`
	LastUpdatedAt time.Time  `json:"last_updated_at" example:"2026-01-01T00:25:00Z"`
	CreatedAt     time.Time  `json:"created_at" example:"2026-01-01T00:25:00Z"`
	UpdatedAt     time.Time  `json:"updated_at" example:"2026-01-01T00:25:00Z"`
}

type EntitlementList struct {
	Data []Entitlement `json:"data"`
}

type OfferRequest struct {
	FreeTrial string `json:"free_trial" example:"NONE" enums:"ACTIVE,INACTIVE,NONE,UNKNOWN"`
}

type UpsertEntitlementRequest struct {
	Products string       `json:"products" example:"com.partner.product1"`
	Offer    OfferRequest `json:"offer"`
}
