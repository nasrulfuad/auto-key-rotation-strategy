package entitlement

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type GetEntitlementInput struct {
	Token         domain.VerifiedToken
	SubjectID     string
	EntitlementID string
}

type GetEntitlementInteractor struct {
	entitlementReader repository.EntitlementReader

	peerProvider string
}

func NewGetEntitlementInteractor(
	entitlementReader repository.EntitlementReader,
	peerProvider string,
	opts ...Option,
) *GetEntitlementInteractor {
	return &GetEntitlementInteractor{entitlementReader: entitlementReader, peerProvider: peerProvider}
}

func (i *GetEntitlementInteractor) Execute(ctx context.Context, in GetEntitlementInput) (domain.Entitlement, error) {
	if err := ensureProvider(in.Token, i.peerProvider); err != nil {
		return domain.Entitlement{}, err
	}
	if err := domain.ValidateSubjectID(in.SubjectID); err != nil {
		return domain.Entitlement{}, err
	}
	if err := domain.ValidateEntitlementID(in.EntitlementID); err != nil {
		return domain.Entitlement{}, err
	}
	return i.entitlementReader.FindBySubjectAndEntitlement(ctx, in.Token.Claims.Provider, in.SubjectID, in.EntitlementID)
}
