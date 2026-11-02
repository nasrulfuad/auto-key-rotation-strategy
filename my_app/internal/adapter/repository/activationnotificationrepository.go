package repository

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
)

// ActivationNotificationReader reads the activation notifications received from partner_app.
type ActivationNotificationReader interface {
	// ListAll returns at most limit notifications, newest received first.
	ListAll(ctx context.Context, limit int) ([]domain.ActivationNotification, error)
}

// ActivationNotificationWriter stores the activation notifications received from partner_app.
type ActivationNotificationWriter interface {
	// Create stores notification and reports whether it was new.
	// A notification whose notification_id is already stored is ignored and returns false.
	Create(ctx context.Context, notification domain.ActivationNotification) (bool, error)
}
