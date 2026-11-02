package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
)

func TestNewEntitlementChangeDefaultsFreeTrialToUnknown(t *testing.T) {
	change, err := domain.NewEntitlementChange("my_app", "SUB1", "ENT1", "com.partner.product1", "", baseTime)

	require.NoError(t, err)
	assert.Equal(t, domain.FreeTrialStatusUnknown, change.FreeTrial)
	assert.Equal(t, baseTime, change.LastUpdatedAt)
}

func TestNewEntitlementChangeRejectsInvalidInput(t *testing.T) {
	_, err := domain.NewEntitlementChange("my_app", "null", "ENT1", "p", "NONE", baseTime)
	assert.ErrorIs(t, err, domain.ErrSubjectIDInvalid)

	_, err = domain.NewEntitlementChange("my_app", "SUB1", "ENT 1", "p", "NONE", baseTime)
	assert.ErrorIs(t, err, domain.ErrEntitlementIDInvalid)

	_, err = domain.NewEntitlementChange("my_app", "SUB1", "ENT1", "", "NONE", baseTime)
	assert.ErrorIs(t, err, domain.Invalid("products", ""))

	_, err = domain.NewEntitlementChange("my_app", "SUB1", "ENT1", strings.Repeat("p", 256), "NONE", baseTime)
	assert.ErrorIs(t, err, domain.Invalid("products", ""))

	_, err = domain.NewEntitlementChange("my_app", "SUB1", "ENT1", "p", "TRIAL", baseTime)
	assert.ErrorIs(t, err, domain.Invalid("offer.free_trial", ""))
}
