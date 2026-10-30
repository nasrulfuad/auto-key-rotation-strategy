package domain_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

func TestNewSubscriptionFromPartnerMirrorsPartnerStatus(t *testing.T) {
	id := uuid.New()
	partnerTerminatedAt := baseTime.Add(-time.Minute)

	active, err := domain.NewSubscriptionFromPartner(id, domain.PartnerEntitlement{
		SubjectID: "SUB1", EntitlementID: "ENT1", Status: domain.EntitlementStatusActive,
	}, "kid-1", baseTime)
	require.NoError(t, err)
	assert.Equal(t, &baseTime, active.ActivatedAt)
	assert.Nil(t, active.TerminatedAt)
	assert.Equal(t, "kid-1", active.SignedWithKeyID)

	terminated, err := domain.NewSubscriptionFromPartner(id, domain.PartnerEntitlement{
		SubjectID: "SUB1", EntitlementID: "ENT1", Status: domain.EntitlementStatusTerminated, TerminatedAt: &partnerTerminatedAt,
	}, "kid-2", baseTime)
	require.NoError(t, err)
	assert.Nil(t, terminated.ActivatedAt)
	assert.Equal(t, &partnerTerminatedAt, terminated.TerminatedAt)

	_, err = domain.NewSubscriptionFromPartner(id, domain.PartnerEntitlement{Status: "PAUSED"}, "kid", baseTime)
	assert.ErrorIs(t, err, domain.ErrInternal)
}
