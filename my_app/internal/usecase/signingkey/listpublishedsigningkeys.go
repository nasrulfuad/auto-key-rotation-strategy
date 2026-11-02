package signingkey

import (
	"context"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/systemclock"
)

type ListPublishedSigningKeysInteractor struct {
	signingKeyReader repository.SigningKeyReader
	clock            systemclock.Clock
}

func NewListPublishedSigningKeysInteractor(signingKeyReader repository.SigningKeyReader, opts ...Option) *ListPublishedSigningKeysInteractor {
	o := newOptions(opts)
	return &ListPublishedSigningKeysInteractor{signingKeyReader: signingKeyReader, clock: o.clock}
}

func (i *ListPublishedSigningKeysInteractor) Execute(ctx context.Context) ([]domain.SigningKey, error) {
	return i.signingKeyReader.ListAllPublished(ctx, i.clock.Now())
}
