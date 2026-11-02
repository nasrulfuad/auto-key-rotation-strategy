package http

import (
	"reflect"
	"testing"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/apidoc"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/apidoc/apidoctest"
)

func TestApidocRequestContract(t *testing.T) {
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.ActivateSubscriptionRequest{}), reflect.TypeOf(activateSubscriptionRequest{}))
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.TerminateSubscriptionRequest{}), reflect.TypeOf(terminateSubscriptionRequest{}))
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.ReceiveActivationNotificationRequest{}), reflect.TypeOf(receiveActivationNotificationRequest{}))
}
