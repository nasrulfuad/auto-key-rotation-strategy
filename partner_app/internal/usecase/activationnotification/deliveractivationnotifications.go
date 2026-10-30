package activationnotification

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/gateway"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock"
)

type DeliverActivationNotificationsInteractor struct {
	activationNotificationReader repository.ActivationNotificationReader
	activationNotificationWriter repository.ActivationNotificationWriter
	activationNotificationSender gateway.ActivationNotificationSender
	tokenSigner                  TokenSigner
	clock                        systemclock.Clock
	logger                       logging.Logger

	audience        domain.TokenAudience
	batchSize       int
	maximumAttempts int
}

func NewDeliverActivationNotificationsInteractor(
	activationNotificationReader repository.ActivationNotificationReader,
	activationNotificationWriter repository.ActivationNotificationWriter,
	activationNotificationSender gateway.ActivationNotificationSender,
	tokenSigner TokenSigner,
	clock systemclock.Clock,
	logger logging.Logger,
	audience domain.TokenAudience,
	batchSize int,
	maximumAttempts int,
) *DeliverActivationNotificationsInteractor {
	if logger == nil {
		logger = logging.NopLogger{}
	}
	return &DeliverActivationNotificationsInteractor{
		activationNotificationReader: activationNotificationReader,
		activationNotificationWriter: activationNotificationWriter,
		activationNotificationSender: activationNotificationSender,
		tokenSigner:                  tokenSigner,
		clock:                        clock,
		logger:                       logger,
		audience:                     audience,
		batchSize:                    batchSize,
		maximumAttempts:              maximumAttempts,
	}
}

func (i *DeliverActivationNotificationsInteractor) Execute(ctx context.Context) error {
	now := i.clock.Now()
	notifications, err := i.activationNotificationReader.ListAllDue(ctx, now, i.batchSize)
	if err != nil {
		return err
	}
	for _, notification := range notifications {
		if err := i.deliver(ctx, notification); err != nil {
			return err
		}
	}
	return nil
}

func (i *DeliverActivationNotificationsInteractor) deliver(ctx context.Context, notification domain.ActivationNotification) error {
	signed, err := i.tokenSigner.Sign(ctx, domain.TokenRequest{
		Audience: i.audience,
		Subject:  notification.SubjectID,
		Provider: notification.Provider,
	})
	if err != nil {
		return err
	}
	sendErr := i.activationNotificationSender.Send(ctx, gateway.ActivationNotificationDelivery{Token: signed.Value, Notification: notification})
	now := i.clock.Now()
	fields := []any{
		"notification_id", notification.ID,
		"event", notification.Event,
		"subject_id", notification.SubjectID,
		"entitlement_id", notification.EntitlementID,
		"kid", signed.KeyID,
		"audience", i.audience,
	}
	if sendErr == nil {
		i.logger.Info(ctx, "activation notification delivered; signed a JWT with the active key, posted it to the peer webhook and marked the notification sent",
			append(fields, "attempts", notification.Attempts+1)...)
		return i.activationNotificationWriter.UpdateSent(ctx, notification.ID, now)
	}
	result := notification.FailAttempt(now, i.maximumAttempts)
	fields = append(fields, "attempts", result.Attempts, "status", result.Status, "error", sendErr)
	if result.Status == domain.ActivationNotificationStatusFailed {
		i.logger.Warn(ctx, "activation notification delivery failed; maximum attempts reached, giving up and marking the notification failed", fields...)
	} else {
		i.logger.Warn(ctx, "activation notification delivery failed; retry scheduled at next_attempt_at",
			append(fields, "next_attempt_at", result.NextAttemptAt)...)
	}
	return i.activationNotificationWriter.UpdateAttemptFailed(ctx, repository.ActivationNotificationAttempt{
		ID:        notification.ID,
		Result:    result,
		LastError: sendErr.Error(),
	})
}
