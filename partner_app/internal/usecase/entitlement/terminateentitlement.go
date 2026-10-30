package entitlement

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type TerminateEntitlementInput struct {
	Token         domain.VerifiedToken
	SubjectID     string
	EntitlementID string
}

type TerminateEntitlementInteractor struct {
	entitlementWriter port.EntitlementWriter
	clock             port.Clock
	uuidGenerator     port.UUIDGenerator
	logger            port.Logger

	peerProvider string
}

func NewTerminateEntitlementInteractor(
	entitlementWriter port.EntitlementWriter,
	clock port.Clock,
	uuidGenerator port.UUIDGenerator,
	logger port.Logger,
	peerProvider string,
) *TerminateEntitlementInteractor {
	if logger == nil {
		logger = port.NopLogger{}
	}
	return &TerminateEntitlementInteractor{
		entitlementWriter: entitlementWriter,
		clock:             clock,
		uuidGenerator:     uuidGenerator,
		logger:            logger,
		peerProvider:      peerProvider,
	}
}

func (i *TerminateEntitlementInteractor) Execute(ctx context.Context, in TerminateEntitlementInput) (domain.Entitlement, error) {
	if err := ensureProvider(in.Token, i.peerProvider); err != nil {
		return domain.Entitlement{}, err
	}
	termination, err := domain.NewEntitlementTermination(in.Token.Claims.Provider, in.SubjectID, in.EntitlementID, in.Token.Claims.IssuedAt)
	if err != nil {
		return domain.Entitlement{}, err
	}
	entitlement, err := i.entitlementWriter.Terminate(ctx, port.EntitlementTerminate{
		ActivationNotificationID: i.uuidGenerator.NewUUID(),
		Termination:              termination,
		Now:                      i.clock.Now(),
	})
	if err != nil {
		return domain.Entitlement{}, err
	}
	i.logger.Info(ctx, "entitlement terminated",
		"provider", entitlement.Provider, "subject_id", entitlement.SubjectID,
		"entitlement_id", entitlement.EntitlementID, "kid", in.Token.KeyID)
	return entitlement, nil
}
