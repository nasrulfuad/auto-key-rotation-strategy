package entitlement

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock"
)

type TerminateEntitlementInput struct {
	Token         domain.VerifiedToken
	SubjectID     string
	EntitlementID string
}

type TerminateEntitlementInteractor struct {
	entitlementWriter repository.EntitlementWriter
	clock             systemclock.Clock
	uuidGenerator     systemclock.UUIDGenerator
	logger            logging.Logger

	peerProvider string
}

func NewTerminateEntitlementInteractor(
	entitlementWriter repository.EntitlementWriter,
	peerProvider string,
	opts ...Option,
) *TerminateEntitlementInteractor {
	o := newOptions(opts)
	return &TerminateEntitlementInteractor{
		entitlementWriter: entitlementWriter,
		clock:             o.clock,
		uuidGenerator:     o.uuidGenerator,
		logger:            o.logger,
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
	entitlement, err := i.entitlementWriter.Terminate(ctx, repository.EntitlementTerminate{
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
