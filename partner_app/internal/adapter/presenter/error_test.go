package presenter_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/presenter"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

func TestErrorPresenterMapsConflictTo409WithItsCodeAndMessage(t *testing.T) {
	err := domain.ErrSigningKeyAlreadyPending

	assert.Equal(t, http.StatusConflict, presenter.Status(err.Kind()))
	assert.Equal(t, "SIGNING_KEY_ALREADY_PENDING", presenter.Code(err))
	assert.Equal(t, err.Message(), presenter.PublicMessage(err))
	assert.Equal(t, "CONFLICT", presenter.Code(domain.ConflictCoded("", "conflict")))
}
