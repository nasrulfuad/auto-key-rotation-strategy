package subscription_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port/mocks"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/subscription"
)

var now = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

type recordingSigner struct {
	requests []domain.TokenRequest
}

func (s *recordingSigner) Sign(_ context.Context, request domain.TokenRequest) (domain.SignedToken, error) {
	s.requests = append(s.requests, request)
	return domain.SignedToken{Value: "signed", KeyID: "kid-1"}, nil
}

func TestActivateSubscriptionCallsPartnerThenStoresLocally(t *testing.T) {
	gateway := mocks.NewMockPartnerEntitlementGateway(t)
	writer := mocks.NewMockSubscriptionWriter(t)
	clock := mocks.NewMockClock(t)
	uuids := mocks.NewMockUUIDGenerator(t)
	signer := &recordingSigner{}
	id := uuid.New()
	change := domain.PartnerEntitlementChange{SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p", FreeTrial: domain.FreeTrialStatusNone}
	partner := domain.PartnerEntitlement{SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p", FreeTrial: domain.FreeTrialStatusNone, Status: domain.EntitlementStatusActive}
	gateway.EXPECT().Upsert(mock.Anything, port.PartnerEntitlementUpsert{Token: "signed", Change: change}).Return(partner, nil)
	clock.EXPECT().Now().Return(now)
	uuids.EXPECT().NewUUID().Return(id)
	writer.EXPECT().Upsert(mock.Anything, domain.Subscription{
		ID: id, SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p", FreeTrial: domain.FreeTrialStatusNone,
		Status: domain.EntitlementStatusActive, SignedWithKeyID: "kid-1", ActivatedAt: &now,
	}, now).Return(domain.Subscription{SubjectID: "SUB1", SignedWithKeyID: "kid-1"}, nil)

	result, err := subscription.NewActivateSubscriptionInteractor(gateway, writer, signer, clock, uuids, nil, "set.aud", "my_app").
		Execute(context.Background(), subscription.ActivateSubscriptionInput{SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p", FreeTrial: "NONE"})

	require.NoError(t, err)
	assert.Equal(t, "kid-1", result.SignedWithKeyID)
	assert.Equal(t, []domain.TokenRequest{{Audience: "set.aud", Provider: "my_app"}}, signer.requests)
}

func TestActivateSubscriptionStoresNothingWhenPartnerRejects(t *testing.T) {
	gateway := mocks.NewMockPartnerEntitlementGateway(t)
	writer := mocks.NewMockSubscriptionWriter(t)
	rejection := domain.PartnerRequestRejected(401, "TOKEN_SIGNING_KEY_NOT_FOUND", "authentication required")
	gateway.EXPECT().Upsert(mock.Anything, mock.Anything).Return(domain.PartnerEntitlement{}, rejection)

	_, err := subscription.NewActivateSubscriptionInteractor(gateway, writer, &recordingSigner{}, mocks.NewMockClock(t), mocks.NewMockUUIDGenerator(t), nil, "set.aud", "my_app").
		Execute(context.Background(), subscription.ActivateSubscriptionInput{SubjectID: "SUB1", EntitlementID: "ENT1", Products: "p"})

	assert.ErrorIs(t, err, rejection)
}

func TestTerminateSubscriptionUsesSetAudienceAndPartnerTerminatedAt(t *testing.T) {
	gateway := mocks.NewMockPartnerEntitlementGateway(t)
	writer := mocks.NewMockSubscriptionWriter(t)
	clock := mocks.NewMockClock(t)
	uuids := mocks.NewMockUUIDGenerator(t)
	signer := &recordingSigner{}
	terminatedAt := now.Add(-time.Second)
	gateway.EXPECT().Terminate(mock.Anything, port.PartnerEntitlementTerminate{Token: "signed", SubjectID: "SUB1", EntitlementID: "ENT1"}).
		Return(domain.PartnerEntitlement{SubjectID: "SUB1", EntitlementID: "ENT1", Status: domain.EntitlementStatusTerminated, TerminatedAt: &terminatedAt}, nil)
	clock.EXPECT().Now().Return(now)
	uuids.EXPECT().NewUUID().Return(uuid.New())
	writer.EXPECT().Upsert(mock.Anything, mock.MatchedBy(func(s domain.Subscription) bool {
		return s.Status == domain.EntitlementStatusTerminated && s.TerminatedAt.Equal(terminatedAt) && s.ActivatedAt == nil
	}), now).Return(domain.Subscription{Status: domain.EntitlementStatusTerminated}, nil)

	_, err := subscription.NewTerminateSubscriptionInteractor(gateway, writer, signer, clock, uuids, nil, "set.aud", "my_app").
		Execute(context.Background(), subscription.TerminateSubscriptionInput{SubjectID: "SUB1", EntitlementID: "ENT1"})

	require.NoError(t, err)
	assert.Equal(t, []domain.TokenRequest{{Audience: "set.aud", Provider: "my_app"}}, signer.requests)
}

func TestTerminateSubscriptionRejectsInvalidIDsWithoutSigning(t *testing.T) {
	signer := &recordingSigner{}

	_, err := subscription.NewTerminateSubscriptionInteractor(mocks.NewMockPartnerEntitlementGateway(t), mocks.NewMockSubscriptionWriter(t), signer,
		mocks.NewMockClock(t), mocks.NewMockUUIDGenerator(t), nil, "set.aud", "my_app").
		Execute(context.Background(), subscription.TerminateSubscriptionInput{SubjectID: "SUB1", EntitlementID: "bad id"})

	assert.ErrorIs(t, err, domain.ErrEntitlementIDInvalid)
	assert.Empty(t, signer.requests)
}
