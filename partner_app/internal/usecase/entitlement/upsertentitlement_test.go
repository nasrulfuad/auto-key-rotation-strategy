package entitlement_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/entitlement"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port/mocks"
)

var now = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

func token(provider string) domain.VerifiedToken {
	return domain.VerifiedToken{KeyID: "kid", Claims: domain.TokenClaims{Provider: provider, IssuedAt: now.Add(-time.Second)}}
}

func TestUpsertEntitlementUsesTokenIssuedAtAsLastUpdated(t *testing.T) {
	writer := mocks.NewMockEntitlementWriter(t)
	clock := mocks.NewMockClock(t)
	uuids := mocks.NewMockUUIDGenerator(t)
	rowID, notificationID := uuid.New(), uuid.New()
	clock.EXPECT().Now().Return(now)
	uuids.EXPECT().NewUUID().Return(rowID).Once()
	uuids.EXPECT().NewUUID().Return(notificationID).Once()
	writer.EXPECT().Upsert(mock.Anything, port.EntitlementUpsert{
		ID:                       rowID,
		ActivationNotificationID: notificationID,
		Change: domain.EntitlementChange{
			Provider: "my_app", SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p",
			FreeTrial: domain.FreeTrialStatusNone, LastUpdatedAt: now.Add(-time.Second),
		},
		Now: now,
	}).Return(domain.Entitlement{Provider: "my_app"}, nil)

	_, err := entitlement.NewUpsertEntitlementInteractor(writer, clock, uuids, nil, "my_app").Execute(context.Background(),
		entitlement.UpsertEntitlementInput{Token: token("my_app"), SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p", FreeTrial: "NONE"})

	require.NoError(t, err)
}

func TestUpsertEntitlementRejectsForeignProvider(t *testing.T) {
	writer := mocks.NewMockEntitlementWriter(t)

	_, err := entitlement.NewUpsertEntitlementInteractor(writer, mocks.NewMockClock(t), mocks.NewMockUUIDGenerator(t), nil, "my_app").
		Execute(context.Background(), entitlement.UpsertEntitlementInput{Token: token("someone_else"), SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p"})

	assert.ErrorIs(t, err, domain.TokenClaimInvalid("provider_not_allowed"))
}

func TestGetEntitlementScopesLookupToTokenProvider(t *testing.T) {
	reader := mocks.NewMockEntitlementReader(t)
	reader.EXPECT().FindBySubjectAndEntitlement(mock.Anything, "my_app", "SUB1", "ENT1").Return(domain.Entitlement{}, domain.ErrEntitlementNotFound)

	_, err := entitlement.NewGetEntitlementInteractor(reader, "my_app").Execute(context.Background(),
		entitlement.GetEntitlementInput{Token: token("my_app"), SubjectID: "SUB1", EntitlementID: "ENT1"})

	assert.ErrorIs(t, err, domain.ErrEntitlementNotFound)
}
