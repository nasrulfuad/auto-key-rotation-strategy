package entitlement

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type GetEntitlementInput struct {
	Token         domain.VerifiedToken
	SubjectID     string
	EntitlementID string
}

type GetEntitlementInteractor struct {
	entitlementReader port.EntitlementReader

	peerProvider string
}

func NewGetEntitlementInteractor(entitlementReader port.EntitlementReader, peerProvider string) *GetEntitlementInteractor {
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
