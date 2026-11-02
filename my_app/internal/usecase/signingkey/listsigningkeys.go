package signingkey

import (
	"context"
	"time"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/repository"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/systemclock"
)

type ListSigningKeysOutput struct {
	Now  time.Time
	Keys []domain.SigningKey
}

type ListSigningKeysInteractor struct {
	signingKeyReader repository.SigningKeyReader
	clock            systemclock.Clock
}

func NewListSigningKeysInteractor(signingKeyReader repository.SigningKeyReader, opts ...Option) *ListSigningKeysInteractor {
	o := newOptions(opts)
	return &ListSigningKeysInteractor{signingKeyReader: signingKeyReader, clock: o.clock}
}

func (i *ListSigningKeysInteractor) Execute(ctx context.Context) (ListSigningKeysOutput, error) {
	now := i.clock.Now()
	keys, err := i.signingKeyReader.ListAll(ctx)
	if err != nil {
		return ListSigningKeysOutput{}, err
	}
	return ListSigningKeysOutput{Now: now, Keys: keys}, nil
}
