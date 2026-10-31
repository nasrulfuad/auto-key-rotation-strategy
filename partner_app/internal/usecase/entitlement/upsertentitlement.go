package entitlement

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock"
)

type UpsertEntitlementInput struct {
	Token         domain.VerifiedToken
	SubjectID     string
	EntitlementID string
	Products      string
	FreeTrial     string
}

type UpsertEntitlementInteractor struct {
	entitlementWriter repository.EntitlementWriter
	clock             systemclock.Clock
	uuidGenerator     systemclock.UUIDGenerator
	logger            logging.Logger

	peerProvider string
}

func NewUpsertEntitlementInteractor(
	entitlementWriter repository.EntitlementWriter,
	peerProvider string,
	opts ...Option,
) *UpsertEntitlementInteractor {
	o := newOptions(opts)
	return &UpsertEntitlementInteractor{
		entitlementWriter: entitlementWriter,
		clock:             o.clock,
		uuidGenerator:     o.uuidGenerator,
		logger:            o.logger,
		peerProvider:      peerProvider,
	}
}

func (i *UpsertEntitlementInteractor) Execute(ctx context.Context, in UpsertEntitlementInput) (domain.Entitlement, error) {
	if err := ensureProvider(in.Token, i.peerProvider); err != nil {
		return domain.Entitlement{}, err
	}
	change, err := domain.NewEntitlementChange(in.Token.Claims.Provider, in.SubjectID, in.EntitlementID, in.Products, in.FreeTrial, in.Token.Claims.IssuedAt)
	if err != nil {
		return domain.Entitlement{}, err
	}
	entitlement, err := i.entitlementWriter.Upsert(ctx, repository.EntitlementUpsert{
		ID:                       i.uuidGenerator.NewUUID(),
		ActivationNotificationID: i.uuidGenerator.NewUUID(),
		Change:                   change,
		Now:                      i.clock.Now(),
	})
	if err != nil {
		return domain.Entitlement{}, err
	}
	i.logger.Info(ctx, "entitlement upserted",
		"provider", entitlement.Provider, "subject_id", entitlement.SubjectID,
		"entitlement_id", entitlement.EntitlementID, "kid", in.Token.KeyID)
	return entitlement, nil
}
