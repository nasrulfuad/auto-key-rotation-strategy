package domain_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

func TestValidateSubjectID(t *testing.T) {
	cases := map[string]bool{
		"MOO1664367665504":       true,
		"a-b_c.d=e":              true,
		strings.Repeat("a", 254): true,
		strings.Repeat("a", 255): false,
		"":                       false,
		"has space":              false,
		"slash/value":            false,
		"percent%20":             false,
		"null":                   false,
		"subNULL1":               false,
		"undefined":              false,
		"nil":                    false,
		"EMPTY":                  false,
	}
	for value, valid := range cases {
		err := domain.ValidateSubjectID(value)
		if valid {
			assert.NoError(t, err, value)
		} else {
			assert.ErrorIs(t, err, domain.ErrSubjectIDInvalid, value)
		}
	}
}

func TestValidateEntitlementIDUsesTheSameRules(t *testing.T) {
	assert.NoError(t, domain.ValidateEntitlementID("ENT1664367665579"))
	assert.ErrorIs(t, domain.ValidateEntitlementID("ent nil"), domain.ErrEntitlementIDInvalid)
}
