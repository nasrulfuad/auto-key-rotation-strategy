package subscription

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

type TerminateSubscriptionInput struct {
	SubjectID     string
	EntitlementID string
}

type TerminateSubscriptionInteractor struct {
	partnerEntitlementGateway port.PartnerEntitlementGateway
	subscriptionWriter        port.SubscriptionWriter
	tokenSigner               TokenSigner
	clock                     port.Clock
	uuidGenerator             port.UUIDGenerator
	logger                    port.Logger

	audience domain.TokenAudience
	provider string
}

func NewTerminateSubscriptionInteractor(
	partnerEntitlementGateway port.PartnerEntitlementGateway,
	subscriptionWriter port.SubscriptionWriter,
	tokenSigner TokenSigner,
	clock port.Clock,
	uuidGenerator port.UUIDGenerator,
	logger port.Logger,
	audience domain.TokenAudience,
	provider string,
) *TerminateSubscriptionInteractor {
	if logger == nil {
		logger = port.NopLogger{}
	}
	return &TerminateSubscriptionInteractor{
		partnerEntitlementGateway: partnerEntitlementGateway,
		subscriptionWriter:        subscriptionWriter,
		tokenSigner:               tokenSigner,
		clock:                     clock,
		uuidGenerator:             uuidGenerator,
		logger:                    logger,
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
	entitlement, err := i.partnerEntitlementGateway.Terminate(ctx, port.PartnerEntitlementTerminate{
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
