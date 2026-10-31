package signingkey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/signingkey"
)

func TestListSigningKeysReturnsEveryKeyWithTheInstantItWasListedAt(t *testing.T) {
	reader := repositorymocks.NewMockSigningKeyReader(t)
	clock := systemclockmocks.NewMockClock(t)
	keys := []domain.SigningKey{{KeyID: "inactive"}, {KeyID: "active"}, {KeyID: "pending"}}
	clock.EXPECT().Now().Return(now)
	reader.EXPECT().ListAll(mock.Anything).Return(keys, nil)

	out, err := signingkey.NewListSigningKeysInteractor(reader, signingkey.WithClock(clock)).Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, signingkey.ListSigningKeysOutput{Now: now, Keys: keys}, out)
}

func TestListSigningKeysReturnsReaderError(t *testing.T) {
	reader := repositorymocks.NewMockSigningKeyReader(t)
	clock := systemclockmocks.NewMockClock(t)
	readErr := domain.DB(errors.New("connection reset"))
	clock.EXPECT().Now().Return(now)
	reader.EXPECT().ListAll(mock.Anything).Return(nil, readErr)

	_, err := signingkey.NewListSigningKeysInteractor(reader, signingkey.WithClock(clock)).Execute(context.Background())

	require.ErrorIs(t, err, readErr)
}
