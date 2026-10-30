package subscription

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

const maximumSubscriptions = 100

type ListSubscriptionsInteractor struct {
	subscriptionReader port.SubscriptionReader
}

func NewListSubscriptionsInteractor(subscriptionReader port.SubscriptionReader) *ListSubscriptionsInteractor {
	return &ListSubscriptionsInteractor{subscriptionReader: subscriptionReader}
}

func (i *ListSubscriptionsInteractor) Execute(ctx context.Context) ([]domain.Subscription, error) {
	return i.subscriptionReader.ListAll(ctx, maximumSubscriptions)
}
