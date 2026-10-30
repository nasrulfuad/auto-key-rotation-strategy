package apidoc

import "time"

type Subscription struct {
	SubjectID       string     `json:"subject_id" example:"SUB001"`
	EntitlementID   string     `json:"entitlement_id" example:"ENT001"`
	Products        string     `json:"products" example:"com.partner.product1"`
	FreeTrial       string     `json:"free_trial" example:"NONE" enums:"ACTIVE,INACTIVE,NONE,UNKNOWN"`
	Status          string     `json:"status" example:"ACTIVE" enums:"ACTIVE,TERMINATED"`
	SignedWithKeyID string     `json:"signed_with_key_id" example:"NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"`
	ActivatedAt     *time.Time `json:"activated_at"`
	TerminatedAt    *time.Time `json:"terminated_at"`
	CreatedAt       time.Time  `json:"created_at" example:"2026-01-01T00:25:00Z"`
	UpdatedAt       time.Time  `json:"updated_at" example:"2026-01-01T00:25:00Z"`
}

type SubscriptionList struct {
	Data []Subscription `json:"data"`
}

type Offer struct {
	FreeTrial string `json:"free_trial" example:"NONE" enums:"ACTIVE,INACTIVE,NONE,UNKNOWN"`
}

type ActivateSubscriptionRequest struct {
	SubjectID     string `json:"subject_id" example:"SUB001"`
	EntitlementID string `json:"entitlement_id" example:"ENT001"`
	Products      string `json:"products" example:"com.partner.product1"`
	Offer         Offer  `json:"offer"`
}

type TerminateSubscriptionRequest struct {
	SubjectID     string `json:"subject_id" example:"SUB001"`
	EntitlementID string `json:"entitlement_id" example:"ENT001"`
}
