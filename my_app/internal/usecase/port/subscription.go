package port

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type SubscriptionReader interface {
	ListAll(ctx context.Context, limit int) ([]domain.Subscription, error)
}

type SubscriptionWriter interface {
	Upsert(ctx context.Context, subscription domain.Subscription, now time.Time) (domain.Subscription, error)
}
