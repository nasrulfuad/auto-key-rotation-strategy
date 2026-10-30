package activationnotification

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

const maximumActivationNotifications = 100

type ListActivationNotificationsInteractor struct {
	activationNotificationReader port.ActivationNotificationReader
}

func NewListActivationNotificationsInteractor(activationNotificationReader port.ActivationNotificationReader) *ListActivationNotificationsInteractor {
	return &ListActivationNotificationsInteractor{activationNotificationReader: activationNotificationReader}
}

func (i *ListActivationNotificationsInteractor) Execute(ctx context.Context) ([]domain.ActivationNotification, error) {
	return i.activationNotificationReader.ListAll(ctx, maximumActivationNotifications)
}
