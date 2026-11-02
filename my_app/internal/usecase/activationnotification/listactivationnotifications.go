package activationnotification

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
)

const maximumActivationNotifications = 100

type ListActivationNotificationsInteractor struct {
	activationNotificationReader repository.ActivationNotificationReader
}

func NewListActivationNotificationsInteractor(activationNotificationReader repository.ActivationNotificationReader, opts ...Option) *ListActivationNotificationsInteractor {
	return &ListActivationNotificationsInteractor{activationNotificationReader: activationNotificationReader}
}

func (i *ListActivationNotificationsInteractor) Execute(ctx context.Context) ([]domain.ActivationNotification, error) {
	return i.activationNotificationReader.ListAll(ctx, maximumActivationNotifications)
}
