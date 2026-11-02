package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/store"
)

type SubscriptionRepository struct {
	queries store.Querier
}

var (
	_ repository.SubscriptionReader = (*SubscriptionRepository)(nil)
	_ repository.SubscriptionWriter = (*SubscriptionRepository)(nil)
)

func NewSubscriptionRepository(db *sql.DB, opts ...Option) *SubscriptionRepository {
	return &SubscriptionRepository{queries: newOptions(db, opts).querier}
}

func (r *SubscriptionRepository) Upsert(ctx context.Context, subscription domain.Subscription, now time.Time) (domain.Subscription, error) {
	row, err := r.queries.UpsertSubscription(ctx, store.UpsertSubscriptionParams{
		ID:              subscription.ID,
		SubjectID:       subscription.SubjectID,
		EntitlementID:   subscription.EntitlementID,
		Products:        subscription.Products,
		FreeTrial:       string(subscription.FreeTrial),
		Status:          string(subscription.Status),
		SignedWithKeyID: subscription.SignedWithKeyID,
		ActivatedAt:     subscription.ActivatedAt,
		TerminatedAt:    subscription.TerminatedAt,
		Now:             now,
	})
	if err != nil {
		return domain.Subscription{}, domain.DB(err)
	}
	return toSubscription(row), nil
}

func (r *SubscriptionRepository) ListAll(ctx context.Context, limit int) ([]domain.Subscription, error) {
	rows, err := r.queries.ListAllSubscriptions(ctx, int32(limit))
	if err != nil {
		return nil, domain.DB(err)
	}
	subscriptions := make([]domain.Subscription, 0, len(rows))
	for _, row := range rows {
		subscriptions = append(subscriptions, toSubscription(row))
	}
	return subscriptions, nil
}

func toSubscription(row store.Subscription) domain.Subscription {
	return domain.Subscription{
		ID:              row.ID,
		SubjectID:       row.SubjectID,
		EntitlementID:   row.EntitlementID,
		Products:        row.Products,
		FreeTrial:       domain.FreeTrialStatus(row.FreeTrial),
		Status:          domain.EntitlementStatus(row.Status),
		SignedWithKeyID: row.SignedWithKeyID,
		ActivatedAt:     toUTC(row.ActivatedAt),
		TerminatedAt:    toUTC(row.TerminatedAt),
		CreatedAt:       row.CreatedAt.UTC(),
		UpdatedAt:       row.UpdatedAt.UTC(),
	}
}

func toUTC(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}
