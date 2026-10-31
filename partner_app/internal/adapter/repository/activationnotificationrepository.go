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

// ActivationNotificationReader reads the activation notifications queued for delivery to my_app.
type ActivationNotificationReader interface {
	// ListAllDue returns at most limit pending notifications whose next_attempt_at is at or before now,
	// oldest due first, joined with their entitlement.
	ListAllDue(ctx context.Context, now time.Time, limit int) ([]domain.ActivationNotification, error)
}

// ActivationNotificationWriter records the outcome of each delivery attempt.
type ActivationNotificationWriter interface {
	// UpdateSent marks notification id as sent at sentAt and clears its last error.
	UpdateSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error

	// UpdateAttemptFailed stores the status, attempt count, next attempt time and error of a failed delivery.
	UpdateAttemptFailed(ctx context.Context, in ActivationNotificationAttempt) error
}
