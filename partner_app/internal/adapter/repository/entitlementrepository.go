package repository

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

// EntitlementReader reads the entitlements granted through a peer provider.
type EntitlementReader interface {
	// FindBySubjectAndEntitlement returns one entitlement, or domain.ErrEntitlementNotFound.
	FindBySubjectAndEntitlement(ctx context.Context, provider, subjectID, entitlementID string) (domain.Entitlement, error)

	// ListAllBySubject returns every entitlement of subjectID for provider, ordered by entitlement_id.
	ListAllBySubject(ctx context.Context, provider, subjectID string) ([]domain.Entitlement, error)
}

// EntitlementWriter changes entitlements and queues an activation notification for each change in the same transaction.
type EntitlementWriter interface {
	// Upsert creates or reactivates the entitlement in in.Change and queues an activation notification.
	// It returns domain.ErrEntitlementStaleUpdate when the stored entitlement was last updated after in.Change.LastUpdatedAt.
	Upsert(ctx context.Context, in EntitlementUpsert) (domain.Entitlement, error)

	// Terminate terminates the entitlement in in.Termination and queues an activation notification.
	// It returns domain.ErrEntitlementNotFound when the entitlement does not exist,
	// and domain.ErrEntitlementStaleUpdate when it was last updated after in.Termination.LastUpdatedAt.
	Terminate(ctx context.Context, in EntitlementTerminate) (domain.Entitlement, error)
}
