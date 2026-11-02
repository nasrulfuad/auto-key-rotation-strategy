package subscription_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/subscription"
)

func TestListSubscriptionsReturnsAtMostOneHundredSubscriptions(t *testing.T) {
	reader := repositorymocks.NewMockSubscriptionReader(t)
	subscriptions := []domain.Subscription{{SubjectID: "SUB1", EntitlementID: "ENT1", SignedWithKeyID: "my-kid"}}
	reader.EXPECT().ListAll(mock.Anything, 100).Return(subscriptions, nil)

	result, err := subscription.NewListSubscriptionsInteractor(reader).Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, subscriptions, result)
}

func TestListSubscriptionsReturnsReaderError(t *testing.T) {
	reader := repositorymocks.NewMockSubscriptionReader(t)
	readErr := domain.DB(errors.New("connection reset"))
	reader.EXPECT().ListAll(mock.Anything, 100).Return(nil, readErr)

	_, err := subscription.NewListSubscriptionsInteractor(reader).Execute(context.Background())

	require.ErrorIs(t, err, readErr)
}
