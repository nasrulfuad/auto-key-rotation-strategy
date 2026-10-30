package presenter

import (
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type SubscriptionPresenter struct{}

func NewSubscriptionPresenter() *SubscriptionPresenter { return &SubscriptionPresenter{} }

type subscriptionBody struct {
	SubjectID       string     `json:"subject_id"`
	EntitlementID   string     `json:"entitlement_id"`
	Products        string     `json:"products"`
	FreeTrial       string     `json:"free_trial"`
	Status          string     `json:"status"`
	SignedWithKeyID string     `json:"signed_with_key_id"`
	ActivatedAt     *time.Time `json:"activated_at"`
	TerminatedAt    *time.Time `json:"terminated_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type subscriptionListBody struct {
	Data []subscriptionBody `json:"data"`
}

func (p *SubscriptionPresenter) GetSubscription(subscription domain.Subscription) subscriptionBody {
	return subscriptionBody{
		SubjectID:       subscription.SubjectID,
		EntitlementID:   subscription.EntitlementID,
		Products:        subscription.Products,
		FreeTrial:       string(subscription.FreeTrial),
		Status:          string(subscription.Status),
		SignedWithKeyID: subscription.SignedWithKeyID,
		ActivatedAt:     subscription.ActivatedAt,
		TerminatedAt:    subscription.TerminatedAt,
		CreatedAt:       subscription.CreatedAt,
		UpdatedAt:       subscription.UpdatedAt,
	}
}

func (p *SubscriptionPresenter) ListSubscriptions(subscriptions []domain.Subscription) subscriptionListBody {
	body := subscriptionListBody{Data: make([]subscriptionBody, 0, len(subscriptions))}
	for _, subscription := range subscriptions {
		body.Data = append(body.Data, p.GetSubscription(subscription))
	}
	return body
}
