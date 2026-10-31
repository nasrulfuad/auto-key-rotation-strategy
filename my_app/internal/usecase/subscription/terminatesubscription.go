package subscription

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/gateway"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/systemclock"
)

type TerminateSubscriptionInput struct {
	SubjectID     string
	EntitlementID string
}

type TerminateSubscriptionInteractor struct {
	partnerEntitlementGateway gateway.PartnerEntitlementGateway
	subscriptionWriter        repository.SubscriptionWriter
	tokenSigner               TokenSigner
	clock                     systemclock.Clock
	uuidGenerator             systemclock.UUIDGenerator
	logger                    logging.Logger

	audience domain.TokenAudience
	provider string
}

func NewTerminateSubscriptionInteractor(
	partnerEntitlementGateway gateway.PartnerEntitlementGateway,
	subscriptionWriter repository.SubscriptionWriter,
	tokenSigner TokenSigner,
	audience domain.TokenAudience,
	provider string,
	opts ...Option,
) *TerminateSubscriptionInteractor {
	o := newOptions(opts)
	return &TerminateSubscriptionInteractor{
		partnerEntitlementGateway: partnerEntitlementGateway,
		subscriptionWriter:        subscriptionWriter,
		tokenSigner:               tokenSigner,
		clock:                     o.clock,
		uuidGenerator:             o.uuidGenerator,
		logger:                    o.logger,
		audience:                  audience,
		provider:                  provider,
	}
}

func (i *TerminateSubscriptionInteractor) Execute(ctx context.Context, in TerminateSubscriptionInput) (domain.Subscription, error) {
	if err := domain.ValidateEntitlementReference(in.SubjectID, in.EntitlementID); err != nil {
		return domain.Subscription{}, err
	}
	signed, err := i.tokenSigner.Sign(ctx, domain.TokenRequest{Audience: i.audience, Provider: i.provider})
	if err != nil {
		return domain.Subscription{}, err
	}
	entitlement, err := i.partnerEntitlementGateway.Terminate(ctx, gateway.PartnerEntitlementTerminate{
		Token:         signed.Value,
		SubjectID:     in.SubjectID,
		EntitlementID: in.EntitlementID,
	})
	if err != nil {
		i.logger.Warn(ctx, "partner_app rejected the termination", "kid", signed.KeyID, "error", err)
		return domain.Subscription{}, err
	}
	now := i.clock.Now()
	subscription, err := domain.NewSubscriptionFromPartner(i.uuidGenerator.NewUUID(), entitlement, signed.KeyID, now)
	if err != nil {
		return domain.Subscription{}, err
	}
	saved, err := i.subscriptionWriter.Upsert(ctx, subscription, now)
	if err != nil {
		return domain.Subscription{}, err
	}
	i.logger.Info(ctx, "subscription terminated", "kid", signed.KeyID, "subject_id", saved.SubjectID, "entitlement_id", saved.EntitlementID)
	return saved, nil
}
