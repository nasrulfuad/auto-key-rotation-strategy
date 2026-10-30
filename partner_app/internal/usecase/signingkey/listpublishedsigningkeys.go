package signingkey

import (
	"context"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/port"
)

type ListPublishedSigningKeysInteractor struct {
	signingKeyReader port.SigningKeyReader
	clock            port.Clock
}

func NewListPublishedSigningKeysInteractor(signingKeyReader port.SigningKeyReader, clock port.Clock) *ListPublishedSigningKeysInteractor {
	return &ListPublishedSigningKeysInteractor{signingKeyReader: signingKeyReader, clock: clock}
}

func (i *ListPublishedSigningKeysInteractor) Execute(ctx context.Context) ([]domain.SigningKey, error) {
	return i.signingKeyReader.ListAllPublished(ctx, i.clock.Now())
}
