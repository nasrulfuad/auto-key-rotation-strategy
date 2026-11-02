package entitlement_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/entitlement"
)

func TestTerminateEntitlementUsesTokenIssuedAtAndQueuesNotification(t *testing.T) {
	writer := repositorymocks.NewMockEntitlementWriter(t)
	clock := systemclockmocks.NewMockClock(t)
	uuids := systemclockmocks.NewMockUUIDGenerator(t)
	notificationID := uuid.New()
	clock.EXPECT().Now().Return(now)
	uuids.EXPECT().NewUUID().Return(notificationID)
	writer.EXPECT().Terminate(mock.Anything, repository.EntitlementTerminate{
		ActivationNotificationID: notificationID,
		Termination:              domain.EntitlementTermination{Provider: "my_app", SubjectID: "SUB1", EntitlementID: "ENT1", LastUpdatedAt: token("my_app").Claims.IssuedAt},
		Now:                      now,
	}).Return(domain.Entitlement{Status: domain.EntitlementStatusTerminated}, nil)

	result, err := entitlement.NewTerminateEntitlementInteractor(writer, "my_app", entitlement.WithClock(clock), entitlement.WithUUIDGenerator(uuids)).Execute(context.Background(),
		entitlement.TerminateEntitlementInput{Token: token("my_app"), SubjectID: "SUB1", EntitlementID: "ENT1"})

	require.NoError(t, err)
	assert.Equal(t, domain.EntitlementStatusTerminated, result.Status)
}

func TestTerminateEntitlementRejectsForeignProviderAndInvalidIDs(t *testing.T) {
	uc := entitlement.NewTerminateEntitlementInteractor(repositorymocks.NewMockEntitlementWriter(t), "my_app", entitlement.WithClock(systemclockmocks.NewMockClock(t)), entitlement.WithUUIDGenerator(systemclockmocks.NewMockUUIDGenerator(t)))

	_, err := uc.Execute(context.Background(), entitlement.TerminateEntitlementInput{Token: token("other"), SubjectID: "SUB1", EntitlementID: "ENT1"})
	assert.ErrorIs(t, err, domain.TokenClaimInvalid("provider_not_allowed"))

	_, err = uc.Execute(context.Background(), entitlement.TerminateEntitlementInput{Token: token("my_app"), SubjectID: "null", EntitlementID: "ENT1"})
	assert.ErrorIs(t, err, domain.ErrSubjectIDInvalid)
}
