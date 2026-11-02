package subscription

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
)

const maximumSubscriptions = 100

type ListSubscriptionsInteractor struct {
	subscriptionReader repository.SubscriptionReader
}

func NewListSubscriptionsInteractor(subscriptionReader repository.SubscriptionReader, opts ...Option) *ListSubscriptionsInteractor {
	return &ListSubscriptionsInteractor{subscriptionReader: subscriptionReader}
}

func (i *ListSubscriptionsInteractor) Execute(ctx context.Context) ([]domain.Subscription, error) {
	return i.subscriptionReader.ListAll(ctx, maximumSubscriptions)
}
