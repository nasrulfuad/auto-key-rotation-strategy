package activationnotification

import (
	"context"

	"github.com/google/uuid"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

type ReceiveActivationNotificationInput struct {
	Token          domain.VerifiedToken
	NotificationID uuid.UUID
	Event          string
	Provider       string
	SubjectID      string
	EntitlementID  string
	Products       string
	FreeTrial      string
}

type ReceiveActivationNotificationInteractor struct {
	activationNotificationWriter port.ActivationNotificationWriter
	clock                        port.Clock
	uuidGenerator                port.UUIDGenerator
	logger                       port.Logger
}

func NewReceiveActivationNotificationInteractor(
	activationNotificationWriter port.ActivationNotificationWriter,
	clock port.Clock,
	uuidGenerator port.UUIDGenerator,
	logger port.Logger,
) *ReceiveActivationNotificationInteractor {
	if logger == nil {
		logger = port.NopLogger{}
	}
	return &ReceiveActivationNotificationInteractor{
		activationNotificationWriter: activationNotificationWriter,
		clock:                        clock,
		uuidGenerator:                uuidGenerator,
		logger:                       logger,
	}
}

func (i *ReceiveActivationNotificationInteractor) Execute(ctx context.Context, in ReceiveActivationNotificationInput) error {
	notification := domain.ActivationNotification{
		ID:             i.uuidGenerator.NewUUID(),
		NotificationID: in.NotificationID,
		KeyID:          in.Token.KeyID,
		Issuer:         in.Token.Claims.Issuer,
		Event:          domain.ActivationNotificationEvent(in.Event),
		Provider:       in.Provider,
		SubjectID:      in.SubjectID,
		EntitlementID:  in.EntitlementID,
		Products:       in.Products,
		FreeTrial:      domain.FreeTrialStatus(in.FreeTrial),
		ReceivedAt:     i.clock.Now(),
	}
	if err := notification.Validate(in.Token); err != nil {
		return err
	}
	created, err := i.activationNotificationWriter.Create(ctx, notification)
	if err != nil {
		return err
	}
	if !created {
		i.logger.Info(ctx, "duplicate activation notification ignored", "notification_id", notification.NotificationID, "kid", notification.KeyID)
		return nil
	}
	i.logger.Info(ctx, "activation notification received",
		"notification_id", notification.NotificationID, "subject_id", notification.SubjectID,
		"entitlement_id", notification.EntitlementID, "kid", notification.KeyID)
	return nil
}
