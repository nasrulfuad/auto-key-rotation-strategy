package http

import (
	"reflect"
	"testing"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/apidoc"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/apidoc/apidoctest"
)

func TestApidocRequestContract(t *testing.T) {
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.UpsertEntitlementRequest{}), reflect.TypeOf(upsertEntitlementRequest{}))
}
