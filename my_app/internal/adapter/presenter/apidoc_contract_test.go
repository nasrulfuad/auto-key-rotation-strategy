package presenter

import (
	"reflect"
	"testing"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/apidoc"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/apidoc/apidoctest"
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
		{"PeerSigningKeyList/peerSigningKeyListBody", reflect.TypeOf(apidoc.PeerSigningKeyList{}), reflect.TypeOf(peerSigningKeyListBody{})},
		{"Subscription/subscriptionBody", reflect.TypeOf(apidoc.Subscription{}), reflect.TypeOf(subscriptionBody{})},
		{"SubscriptionList/subscriptionListBody", reflect.TypeOf(apidoc.SubscriptionList{}), reflect.TypeOf(subscriptionListBody{})},
		{"ActivationNotificationList/activationNotificationListBody", reflect.TypeOf(apidoc.ActivationNotificationList{}), reflect.TypeOf(activationNotificationListBody{})},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			apidoctest.AssertShapeMatches(t, tc.mirror, tc.real)
		})
	}
}
