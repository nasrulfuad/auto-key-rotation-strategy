package activationnotification_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/activationnotification"
)

func TestReceiveActivationNotificationStoresSigningKeyID(t *testing.T) {
	writer := repositorymocks.NewMockActivationNotificationWriter(t)
	clock := systemclockmocks.NewMockClock(t)
	uuids := systemclockmocks.NewMockUUIDGenerator(t)
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	rowID, notificationID := uuid.New(), uuid.New()
	clock.EXPECT().Now().Return(now)
	uuids.EXPECT().NewUUID().Return(rowID)
	writer.EXPECT().Create(mock.Anything, domain.ActivationNotification{
		ID: rowID, NotificationID: notificationID, KeyID: "partner-kid", Issuer: "qa.partner-app.local",
		Event: domain.ActivationNotificationEventEntitlementUpserted, Provider: "my_app",
		SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p", FreeTrial: domain.FreeTrialStatusNone, ReceivedAt: now,
	}).Return(false, nil)

	err := activationnotification.NewReceiveActivationNotificationInteractor(writer, activationnotification.WithClock(clock), activationnotification.WithUUIDGenerator(uuids)).Execute(context.Background(),
		activationnotification.ReceiveActivationNotificationInput{
			Token: domain.VerifiedToken{KeyID: "partner-kid", Claims: domain.TokenClaims{
				Issuer: "qa.partner-app.local", Subject: "SUB1", Provider: "my_app",
			}},
			NotificationID: notificationID, Event: "entitlement_upserted", Provider: "my_app",
			SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p", FreeTrial: "NONE",
		})

	assert.NoError(t, err)
}
