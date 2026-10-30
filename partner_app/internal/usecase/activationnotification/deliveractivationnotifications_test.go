package activationnotification_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/activationnotification"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port/mocks"
)

var now = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

type stubSigner struct{}

func (stubSigner) Sign(_ context.Context, request domain.TokenRequest) (domain.SignedToken, error) {
	return domain.SignedToken{Value: "token-for-" + request.Subject, KeyID: "kid"}, nil
}

func deliveryFields(notification domain.ActivationNotification) []any {
	return []any{
		"notification_id", notification.ID,
		"event", notification.Event,
		"subject_id", notification.SubjectID,
		"entitlement_id", notification.EntitlementID,
		"kid", "kid",
		"audience", domain.TokenAudience("activation-notification.my-app.local"),
	}
}

func TestDeliverActivationNotificationsMarksSentAndSchedulesRetry(t *testing.T) {
	reader := mocks.NewMockActivationNotificationReader(t)
	writer := mocks.NewMockActivationNotificationWriter(t)
	sender := mocks.NewMockActivationNotificationSender(t)
	clock := mocks.NewMockClock(t)
	logger := mocks.NewMockLogger(t)
	delivered := domain.ActivationNotification{ID: uuid.New(), SubjectID: "SUB1", Provider: "my_app"}
	failing := domain.ActivationNotification{ID: uuid.New(), SubjectID: "SUB2", Provider: "my_app", Attempts: 1}
	clock.EXPECT().Now().Return(now)
	reader.EXPECT().ListAllDue(mock.Anything, now, 10).Return([]domain.ActivationNotification{delivered, failing}, nil)
	sender.EXPECT().Send(mock.Anything, port.ActivationNotificationDelivery{Token: "token-for-SUB1", Notification: delivered}).Return(nil)
	sender.EXPECT().Send(mock.Anything, port.ActivationNotificationDelivery{Token: "token-for-SUB2", Notification: failing}).Return(errors.New("status 401"))
	writer.EXPECT().UpdateSent(mock.Anything, delivered.ID, now).Return(nil)
	writer.EXPECT().UpdateAttemptFailed(mock.Anything, port.ActivationNotificationAttempt{
		ID:        failing.ID,
		Result:    domain.ActivationNotificationAttemptResult{Attempts: 2, Status: domain.ActivationNotificationStatusPending, NextAttemptAt: now.Add(10 * time.Second)},
		LastError: "status 401",
	}).Return(nil)
	logger.EXPECT().Info(mock.Anything,
		"activation notification delivered; signed a JWT with the active key, posted it to the peer webhook and marked the notification sent",
		append(deliveryFields(delivered), "attempts", 1)).Return().Once()
	logger.EXPECT().Warn(mock.Anything,
		"activation notification delivery failed; retry scheduled at next_attempt_at",
		append(deliveryFields(failing), "attempts", 2, "status", domain.ActivationNotificationStatusPending, "error", errors.New("status 401"),
			"next_attempt_at", now.Add(10*time.Second))).Return().Once()

	err := activationnotification.NewDeliverActivationNotificationsInteractor(reader, writer, sender, stubSigner{}, clock, logger,
		"activation-notification.my-app.local", 10, 5).Execute(context.Background())

	require.NoError(t, err)
}

func TestDeliverActivationNotificationsGivesUpAfterMaximumAttempts(t *testing.T) {
	reader := mocks.NewMockActivationNotificationReader(t)
	writer := mocks.NewMockActivationNotificationWriter(t)
	sender := mocks.NewMockActivationNotificationSender(t)
	clock := mocks.NewMockClock(t)
	logger := mocks.NewMockLogger(t)
	exhausted := domain.ActivationNotification{ID: uuid.New(), SubjectID: "SUB3", Provider: "my_app", Attempts: 4}
	sendErr := errors.New("status 500")
	clock.EXPECT().Now().Return(now)
	reader.EXPECT().ListAllDue(mock.Anything, now, 10).Return([]domain.ActivationNotification{exhausted}, nil)
	sender.EXPECT().Send(mock.Anything, port.ActivationNotificationDelivery{Token: "token-for-SUB3", Notification: exhausted}).Return(sendErr)
	writer.EXPECT().UpdateAttemptFailed(mock.Anything, port.ActivationNotificationAttempt{
		ID:        exhausted.ID,
		Result:    domain.ActivationNotificationAttemptResult{Attempts: 5, Status: domain.ActivationNotificationStatusFailed, NextAttemptAt: now},
		LastError: "status 500",
	}).Return(nil)
	logger.EXPECT().Warn(mock.Anything,
		"activation notification delivery failed; maximum attempts reached, giving up and marking the notification failed",
		append(deliveryFields(exhausted), "attempts", 5, "status", domain.ActivationNotificationStatusFailed, "error", sendErr)).Return().Once()

	err := activationnotification.NewDeliverActivationNotificationsInteractor(reader, writer, sender, stubSigner{}, clock, logger,
		"activation-notification.my-app.local", 10, 5).Execute(context.Background())

	require.NoError(t, err)
}
