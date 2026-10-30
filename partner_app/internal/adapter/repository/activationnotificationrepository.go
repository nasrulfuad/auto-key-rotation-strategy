package repository

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type ActivationNotificationAttempt struct {
	ID        uuid.UUID
	Result    domain.ActivationNotificationAttemptResult
	LastError string
}

type ActivationNotificationReader interface {
	ListAllDue(ctx context.Context, now time.Time, limit int) ([]domain.ActivationNotification, error)
}

type ActivationNotificationWriter interface {
	UpdateSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
	UpdateAttemptFailed(ctx context.Context, in ActivationNotificationAttempt) error
}
