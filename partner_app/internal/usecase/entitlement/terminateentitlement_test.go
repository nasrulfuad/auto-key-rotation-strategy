package entitlement_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/entitlement"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port/mocks"
)

func TestTerminateEntitlementUsesTokenIssuedAtAndQueuesNotification(t *testing.T) {
	writer := mocks.NewMockEntitlementWriter(t)
	clock := mocks.NewMockClock(t)
	uuids := mocks.NewMockUUIDGenerator(t)
	notificationID := uuid.New()
	clock.EXPECT().Now().Return(now)
	uuids.EXPECT().NewUUID().Return(notificationID)
	writer.EXPECT().Terminate(mock.Anything, port.EntitlementTerminate{
		ActivationNotificationID: notificationID,
		Termination:              domain.EntitlementTermination{Provider: "my_app", SubjectID: "SUB1", EntitlementID: "ENT1", LastUpdatedAt: token("my_app").Claims.IssuedAt},
		Now:                      now,
	}).Return(domain.Entitlement{Status: domain.EntitlementStatusTerminated}, nil)

	result, err := entitlement.NewTerminateEntitlementInteractor(writer, clock, uuids, nil, "my_app").Execute(context.Background(),
		entitlement.TerminateEntitlementInput{Token: token("my_app"), SubjectID: "SUB1", EntitlementID: "ENT1"})

	require.NoError(t, err)
	assert.Equal(t, domain.EntitlementStatusTerminated, result.Status)
}

func TestTerminateEntitlementRejectsForeignProviderAndInvalidIDs(t *testing.T) {
	uc := entitlement.NewTerminateEntitlementInteractor(mocks.NewMockEntitlementWriter(t), mocks.NewMockClock(t), mocks.NewMockUUIDGenerator(t), nil, "my_app")

	_, err := uc.Execute(context.Background(), entitlement.TerminateEntitlementInput{Token: token("other"), SubjectID: "SUB1", EntitlementID: "ENT1"})
	assert.ErrorIs(t, err, domain.TokenClaimInvalid("provider_not_allowed"))

	_, err = uc.Execute(context.Background(), entitlement.TerminateEntitlementInput{Token: token("my_app"), SubjectID: "null", EntitlementID: "ENT1"})
	assert.ErrorIs(t, err, domain.ErrSubjectIDInvalid)
}
