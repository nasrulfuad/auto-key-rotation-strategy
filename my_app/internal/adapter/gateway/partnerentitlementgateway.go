package gateway

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

type PartnerEntitlementUpsert struct {
	Token  string
	Change domain.PartnerEntitlementChange
}

type PartnerEntitlementTerminate struct {
	Token         string
	SubjectID     string
	EntitlementID string
}

// PartnerEntitlementGateway calls the Entitlement API of partner_app with a bearer token signed by my_app.
type PartnerEntitlementGateway interface {
	// Upsert activates or reactivates the entitlement in in.Change and returns partner_app's view of it.
	// A non-2xx answer returns a PARTNER_REQUEST_REJECTED error carrying partner_app's status and code;
	// an unreachable partner_app or an undecodable answer returns PARTNER_UNAVAILABLE.
	Upsert(ctx context.Context, in PartnerEntitlementUpsert) (domain.PartnerEntitlement, error)

	// Terminate terminates the entitlement and returns partner_app's view of it, with the same errors as Upsert.
	Terminate(ctx context.Context, in PartnerEntitlementTerminate) (domain.PartnerEntitlement, error)
}
