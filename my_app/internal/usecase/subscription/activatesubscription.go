package subscription

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/gateway"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/systemclock"
)

type ActivateSubscriptionInput struct {
	SubjectID     string
	EntitlementID string
	Products      string
	FreeTrial     string
}

type ActivateSubscriptionInteractor struct {
	partnerEntitlementGateway gateway.PartnerEntitlementGateway
	subscriptionWriter        repository.SubscriptionWriter
	tokenSigner               TokenSigner
	clock                     systemclock.Clock
	uuidGenerator             systemclock.UUIDGenerator
	logger                    logging.Logger

	audience domain.TokenAudience
	provider string
}

func NewActivateSubscriptionInteractor(
	partnerEntitlementGateway gateway.PartnerEntitlementGateway,
	subscriptionWriter repository.SubscriptionWriter,
	tokenSigner TokenSigner,
	audience domain.TokenAudience,
	provider string,
	opts ...Option,
) *ActivateSubscriptionInteractor {
	o := newOptions(opts)
	return &ActivateSubscriptionInteractor{
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

func (i *ActivateSubscriptionInteractor) Execute(ctx context.Context, in ActivateSubscriptionInput) (domain.Subscription, error) {
	change, err := domain.NewPartnerEntitlementChange(in.SubjectID, in.EntitlementID, in.Products, in.FreeTrial)
	if err != nil {
		return domain.Subscription{}, err
	}
	signed, err := i.tokenSigner.Sign(ctx, domain.TokenRequest{Audience: i.audience, Provider: i.provider})
	if err != nil {
		return domain.Subscription{}, err
	}
	entitlement, err := i.partnerEntitlementGateway.Upsert(ctx, gateway.PartnerEntitlementUpsert{Token: signed.Value, Change: change})
	if err != nil {
		i.logger.Warn(ctx, "partner_app rejected the activation", "kid", signed.KeyID, "error", err)
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
	i.logger.Info(ctx, "subscription activated", "kid", signed.KeyID, "subject_id", saved.SubjectID, "entitlement_id", saved.EntitlementID)
	return saved, nil
}
