package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type EntitlementUpsert struct {
	ID                       uuid.UUID
	ActivationNotificationID uuid.UUID
	Change                   domain.EntitlementChange
	Now                      time.Time
}

type EntitlementTerminate struct {
	ActivationNotificationID uuid.UUID
	Termination              domain.EntitlementTermination
	Now                      time.Time
}

type EntitlementReader interface {
	FindBySubjectAndEntitlement(ctx context.Context, provider, subjectID, entitlementID string) (domain.Entitlement, error)
	ListAllBySubject(ctx context.Context, provider, subjectID string) ([]domain.Entitlement, error)
}

type EntitlementWriter interface {
	Upsert(ctx context.Context, in EntitlementUpsert) (domain.Entitlement, error)
	Terminate(ctx context.Context, in EntitlementTerminate) (domain.Entitlement, error)
}
