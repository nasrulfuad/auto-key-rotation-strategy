package activationnotification_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/activationnotification"
)

func TestListActivationNotificationsReturnsAtMostOneHundredNotifications(t *testing.T) {
	reader := repositorymocks.NewMockActivationNotificationReader(t)
	notifications := []domain.ActivationNotification{{KeyID: "partner-kid-2"}, {KeyID: "partner-kid-1"}}
	reader.EXPECT().ListAll(mock.Anything, 100).Return(notifications, nil)

	result, err := activationnotification.NewListActivationNotificationsInteractor(reader).Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, notifications, result)
}

func TestListActivationNotificationsReturnsReaderError(t *testing.T) {
	reader := repositorymocks.NewMockActivationNotificationReader(t)
	readErr := domain.DB(errors.New("connection reset"))
	reader.EXPECT().ListAll(mock.Anything, 100).Return(nil, readErr)

	_, err := activationnotification.NewListActivationNotificationsInteractor(reader).Execute(context.Background())

	require.ErrorIs(t, err, readErr)
}
