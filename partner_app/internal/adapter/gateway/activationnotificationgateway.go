package gateway

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type ActivationNotificationDelivery struct {
	Token        string
	Notification domain.ActivationNotification
}

type ActivationNotificationSender interface {
	Send(ctx context.Context, in ActivationNotificationDelivery) error
}
