package presenter

import (
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type ActivationNotificationPresenter struct{}

func NewActivationNotificationPresenter(opts ...Option) *ActivationNotificationPresenter {
	return &ActivationNotificationPresenter{}
}

type activationNotificationListBody struct {
	Data []activationNotificationBody `json:"data"`
}

type activationNotificationBody struct {
	NotificationID string    `json:"notification_id"`
	KeyID          string    `json:"key_id"`
	Issuer         string    `json:"issuer"`
	Event          string    `json:"event"`
	Provider       string    `json:"provider"`
	SubjectID      string    `json:"subject_id"`
	EntitlementID  string    `json:"entitlement_id"`
	Products       string    `json:"products"`
	FreeTrial      string    `json:"free_trial"`
	ReceivedAt     time.Time `json:"received_at"`
}

func (p *ActivationNotificationPresenter) ListActivationNotifications(notifications []domain.ActivationNotification) activationNotificationListBody {
	body := activationNotificationListBody{Data: make([]activationNotificationBody, 0, len(notifications))}
	for _, notification := range notifications {
		body.Data = append(body.Data, activationNotificationBody{
			NotificationID: notification.NotificationID.String(),
			KeyID:          notification.KeyID,
			Issuer:         notification.Issuer,
			Event:          string(notification.Event),
			Provider:       notification.Provider,
			SubjectID:      notification.SubjectID,
			EntitlementID:  notification.EntitlementID,
			Products:       notification.Products,
			FreeTrial:      string(notification.FreeTrial),
			ReceivedAt:     notification.ReceivedAt,
		})
	}
	return body
}
