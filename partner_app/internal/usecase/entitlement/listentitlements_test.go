package entitlement_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/entitlement"
)

func TestListEntitlementsListsSubjectEntitlementsOfTheTokenProvider(t *testing.T) {
	reader := repositorymocks.NewMockEntitlementReader(t)
	entitlements := []domain.Entitlement{{SubjectID: "SUB1", EntitlementID: "ENT1", Status: domain.EntitlementStatusActive}}
	reader.EXPECT().ListAllBySubject(mock.Anything, "my_app", "SUB1").Return(entitlements, nil)

	result, err := entitlement.NewListEntitlementsInteractor(reader, "my_app").Execute(context.Background(),
		entitlement.ListEntitlementsInput{Token: token("my_app"), SubjectID: "SUB1"})

	require.NoError(t, err)
	assert.Equal(t, entitlements, result)
}

func TestListEntitlementsRejectsForeignProviderAndInvalidSubjectID(t *testing.T) {
	uc := entitlement.NewListEntitlementsInteractor(repositorymocks.NewMockEntitlementReader(t), "my_app")

	_, err := uc.Execute(context.Background(), entitlement.ListEntitlementsInput{Token: token("other"), SubjectID: "SUB1"})
	assert.ErrorIs(t, err, domain.TokenClaimInvalid("provider_not_allowed"))

	_, err = uc.Execute(context.Background(), entitlement.ListEntitlementsInput{Token: token("my_app"), SubjectID: "null"})
	assert.ErrorIs(t, err, domain.ErrSubjectIDInvalid)
}
