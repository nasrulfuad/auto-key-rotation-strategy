package repository

import (
	"context"
	"time"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
)

// SubscriptionReader reads the local copy of the subscriptions held at partner_app.
type SubscriptionReader interface {
	// ListAll returns at most limit subscriptions, most recently updated first.
	ListAll(ctx context.Context, limit int) ([]domain.Subscription, error)
}

// SubscriptionWriter stores the local copy of the subscriptions held at partner_app.
type SubscriptionWriter interface {
	// Upsert inserts subscription, or updates the row with the same subject_id and entitlement_id, stamps now and returns the stored row.
	Upsert(ctx context.Context, subscription domain.Subscription, now time.Time) (domain.Subscription, error)
}
