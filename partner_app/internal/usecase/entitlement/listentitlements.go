package entitlement

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
)

type ListEntitlementsInput struct {
	Token     domain.VerifiedToken
	SubjectID string
}

type ListEntitlementsInteractor struct {
	entitlementReader repository.EntitlementReader

	peerProvider string
}

func NewListEntitlementsInteractor(
	entitlementReader repository.EntitlementReader,
	peerProvider string,
	opts ...Option,
) *ListEntitlementsInteractor {
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
