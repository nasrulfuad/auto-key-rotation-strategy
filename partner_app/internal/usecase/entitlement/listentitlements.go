package entitlement

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type ListEntitlementsInput struct {
	Token     domain.VerifiedToken
	SubjectID string
}

type ListEntitlementsInteractor struct {
	entitlementReader port.EntitlementReader

	peerProvider string
}

func NewListEntitlementsInteractor(entitlementReader port.EntitlementReader, peerProvider string) *ListEntitlementsInteractor {
	return &ListEntitlementsInteractor{entitlementReader: entitlementReader, peerProvider: peerProvider}
}

func (i *ListEntitlementsInteractor) Execute(ctx context.Context, in ListEntitlementsInput) ([]domain.Entitlement, error) {
	if err := ensureProvider(in.Token, i.peerProvider); err != nil {
		return nil, err
	}
	if err := domain.ValidateSubjectID(in.SubjectID); err != nil {
		return nil, err
	}
	return i.entitlementReader.ListAllBySubject(ctx, in.Token.Claims.Provider, in.SubjectID)
}
