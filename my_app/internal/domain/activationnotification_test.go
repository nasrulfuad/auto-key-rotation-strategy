package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

func TestActivationNotificationValidateBindsBodyToToken(t *testing.T) {
	token := domain.VerifiedToken{Claims: domain.TokenClaims{Subject: "SUB1", Provider: "my_app"}}
	notification := domain.ActivationNotification{
		NotificationID: uuid.New(), Event: domain.ActivationNotificationEventEntitlementUpserted,
		Provider: "my_app", SubjectID: "SUB1", EntitlementID: "ENT1",
	}
	assert.NoError(t, notification.Validate(token))

	forged := notification
	forged.SubjectID = "SUB2"
	assert.ErrorIs(t, forged.Validate(token), domain.TokenClaimInvalid("subject_mismatch"))

	otherProvider := notification
	otherProvider.Provider = "other"
	assert.ErrorIs(t, otherProvider.Validate(token), domain.TokenClaimInvalid("provider_mismatch"))

	unknownEvent := notification
	unknownEvent.Event = "deleted"
	assert.ErrorIs(t, unknownEvent.Validate(token), domain.Invalid("event", ""))
}

func TestNewPartnerEntitlementChangeValidatesBeforeCallingPartner(t *testing.T) {
	_, err := domain.NewPartnerEntitlementChange("undefined", "ENT1", "p", "")
	assert.ErrorIs(t, err, domain.ErrSubjectIDInvalid)

	_, err = domain.NewPartnerEntitlementChange("SUB1", "ENT1", "p", "MAYBE")
	assert.ErrorIs(t, err, domain.Invalid("offer.free_trial", ""))

	change, err := domain.NewPartnerEntitlementChange("SUB1", "ENT1", "p", "")
	assert.NoError(t, err)
	assert.Equal(t, domain.FreeTrialStatus(""), change.FreeTrial)
}
