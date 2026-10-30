package http

import (
	"reflect"
	"testing"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/apidoc"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/apidoc/apidoctest"
)

func TestApidocRequestContract(t *testing.T) {
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.ActivateSubscriptionRequest{}), reflect.TypeOf(activateSubscriptionRequest{}))
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.TerminateSubscriptionRequest{}), reflect.TypeOf(terminateSubscriptionRequest{}))
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.ReceiveActivationNotificationRequest{}), reflect.TypeOf(receiveActivationNotificationRequest{}))
}
