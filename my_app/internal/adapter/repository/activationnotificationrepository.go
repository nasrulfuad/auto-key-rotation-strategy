package repository

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type ActivationNotificationReader interface {
	ListAll(ctx context.Context, limit int) ([]domain.ActivationNotification, error)
}

type ActivationNotificationWriter interface {
	Create(ctx context.Context, notification domain.ActivationNotification) (bool, error)
}
