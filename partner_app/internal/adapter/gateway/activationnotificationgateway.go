package gateway

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
)

type ActivationNotificationDelivery struct {
	Token        string
	Notification domain.ActivationNotification
}

// ActivationNotificationSender delivers activation notifications to the webhook of my_app.
type ActivationNotificationSender interface {
	// Send posts in.Notification with in.Token as bearer token.
	// It returns an error when my_app cannot be reached or answers with a non-2xx status.
	Send(ctx context.Context, in ActivationNotificationDelivery) error
}
