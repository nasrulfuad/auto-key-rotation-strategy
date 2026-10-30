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

type PartnerEntitlementGateway interface {
	Upsert(ctx context.Context, in PartnerEntitlementUpsert) (domain.PartnerEntitlement, error)
	Terminate(ctx context.Context, in PartnerEntitlementTerminate) (domain.PartnerEntitlement, error)
}
