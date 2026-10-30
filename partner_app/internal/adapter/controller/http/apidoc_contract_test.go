package http

import (
	"reflect"
	"testing"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/apidoc"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/apidoc/apidoctest"
)

func TestApidocRequestContract(t *testing.T) {
	apidoctest.AssertShapeMatches(t, reflect.TypeOf(apidoc.UpsertEntitlementRequest{}), reflect.TypeOf(upsertEntitlementRequest{}))
}
