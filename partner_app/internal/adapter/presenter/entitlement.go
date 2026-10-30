package presenter

import (
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type EntitlementPresenter struct{}

func NewEntitlementPresenter() *EntitlementPresenter { return &EntitlementPresenter{} }

type entitlementBody struct {
	Provider      string     `json:"provider"`
	SubjectID     string     `json:"subject_id"`
	EntitlementID string     `json:"entitlement_id"`
	Products      string     `json:"products"`
	Offer         offerBody  `json:"offer"`
	Status        string     `json:"status"`
	TerminatedAt  *time.Time `json:"terminated_at"`
	LastUpdatedAt time.Time  `json:"last_updated_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type offerBody struct {
	FreeTrial string `json:"free_trial"`
}

type entitlementListBody struct {
	Data []entitlementBody `json:"data"`
}

func (p *EntitlementPresenter) GetEntitlement(entitlement domain.Entitlement) entitlementBody {
	return entitlementBody{
		Provider:      entitlement.Provider,
		SubjectID:     entitlement.SubjectID,
		EntitlementID: entitlement.EntitlementID,
		Products:      entitlement.Products,
		Offer:         offerBody{FreeTrial: string(entitlement.FreeTrial)},
		Status:        string(entitlement.Status),
		TerminatedAt:  entitlement.TerminatedAt,
		LastUpdatedAt: entitlement.LastUpdatedAt,
		CreatedAt:     entitlement.CreatedAt,
		UpdatedAt:     entitlement.UpdatedAt,
	}
}

func (p *EntitlementPresenter) ListEntitlements(entitlements []domain.Entitlement) entitlementListBody {
	body := entitlementListBody{Data: make([]entitlementBody, 0, len(entitlements))}
	for _, entitlement := range entitlements {
		body.Data = append(body.Data, p.GetEntitlement(entitlement))
	}
	return body
}
