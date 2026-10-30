package signingkey

import (
	"context"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

type ListSigningKeysOutput struct {
	Now  time.Time
	Keys []domain.SigningKey
}

type ListSigningKeysInteractor struct {
	signingKeyReader port.SigningKeyReader
	clock            port.Clock
}

func NewListSigningKeysInteractor(signingKeyReader port.SigningKeyReader, clock port.Clock) *ListSigningKeysInteractor {
	return &ListSigningKeysInteractor{signingKeyReader: signingKeyReader, clock: clock}
}

func (i *ListSigningKeysInteractor) Execute(ctx context.Context) (ListSigningKeysOutput, error) {
	now := i.clock.Now()
	keys, err := i.signingKeyReader.ListAll(ctx)
	if err != nil {
		return ListSigningKeysOutput{}, err
	}
	return ListSigningKeysOutput{Now: now, Keys: keys}, nil
}
