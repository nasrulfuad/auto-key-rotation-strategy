package signingkey_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/repository/repositorymocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/systemclock/systemclockmocks"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/signingkey"
)

func TestListPublishedSigningKeysAsksReaderForKeysPublishedNow(t *testing.T) {
	reader := repositorymocks.NewMockSigningKeyReader(t)
	clock := systemclockmocks.NewMockClock(t)
	keys := []domain.SigningKey{{KeyID: "active"}, {KeyID: "pending"}}
	clock.EXPECT().Now().Return(now)
	reader.EXPECT().ListAllPublished(mock.Anything, now).Return(keys, nil)

	result, err := signingkey.NewListPublishedSigningKeysInteractor(reader, signingkey.WithClock(clock)).Execute(context.Background())

	require.NoError(t, err)
	assert.Equal(t, keys, result)
}

func TestListPublishedSigningKeysReturnsReaderError(t *testing.T) {
	reader := repositorymocks.NewMockSigningKeyReader(t)
	clock := systemclockmocks.NewMockClock(t)
	readErr := domain.DB(errors.New("connection reset"))
	clock.EXPECT().Now().Return(now)
	reader.EXPECT().ListAllPublished(mock.Anything, now).Return(nil, readErr)

	_, err := signingkey.NewListPublishedSigningKeysInteractor(reader, signingkey.WithClock(clock)).Execute(context.Background())

	require.ErrorIs(t, err, readErr)
}
