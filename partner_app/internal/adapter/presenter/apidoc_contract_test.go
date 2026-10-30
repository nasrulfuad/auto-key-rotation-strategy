package presenter

import (
	"reflect"
	"testing"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/apidoc"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/apidoc/apidoctest"
)

func TestApidocContract(t *testing.T) {
	cases := []struct {
		name   string
		mirror reflect.Type
		real   reflect.Type
	}{
		{"ErrorEnvelope/errorBody", reflect.TypeOf(apidoc.ErrorEnvelope{}), reflect.TypeOf(errorBody{})},
		{"Health/healthBody", reflect.TypeOf(apidoc.Health{}), reflect.TypeOf(healthBody{})},
		{"JSONWebKeySet/jsonWebKeySetBody", reflect.TypeOf(apidoc.JSONWebKeySet{}), reflect.TypeOf(jsonWebKeySetBody{})},
		{"SigningKeyList/signingKeyListBody", reflect.TypeOf(apidoc.SigningKeyList{}), reflect.TypeOf(signingKeyListBody{})},
		{"SigningKeyRotation/signingKeyRotationBody", reflect.TypeOf(apidoc.SigningKeyRotation{}), reflect.TypeOf(signingKeyRotationBody{})},
		{"PeerSigningKeyList/peerSigningKeyListBody", reflect.TypeOf(apidoc.PeerSigningKeyList{}), reflect.TypeOf(peerSigningKeyListBody{})},
		{"Entitlement/entitlementBody", reflect.TypeOf(apidoc.Entitlement{}), reflect.TypeOf(entitlementBody{})},
		{"EntitlementList/entitlementListBody", reflect.TypeOf(apidoc.EntitlementList{}), reflect.TypeOf(entitlementListBody{})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			apidoctest.AssertShapeMatches(t, tc.mirror, tc.real)
		})
	}
}
