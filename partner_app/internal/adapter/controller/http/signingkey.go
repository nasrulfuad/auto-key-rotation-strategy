package http

import (
	"fmt"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/adapter/presenter"
	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/usecase/signingkey"
)

type SigningKeyController struct {
	listPublishedSigningKeys *signingkey.ListPublishedSigningKeysInteractor
	listSigningKeys          *signingkey.ListSigningKeysInteractor
	rotateSigningKey         *signingkey.RotateSigningKeyInteractor
	presenter                *presenter.SigningKeyPresenter

	cacheMaxAge time.Duration
}

func NewSigningKeyController(
	listPublishedSigningKeys *signingkey.ListPublishedSigningKeysInteractor,
	listSigningKeys *signingkey.ListSigningKeysInteractor,
	rotateSigningKey *signingkey.RotateSigningKeyInteractor,
	signingKeyPresenter *presenter.SigningKeyPresenter,
	cacheMaxAge time.Duration,
) *SigningKeyController {
	return &SigningKeyController{
		listPublishedSigningKeys: listPublishedSigningKeys,
		listSigningKeys:          listSigningKeys,
		rotateSigningKey:         rotateSigningKey,
		presenter:                signingKeyPresenter,
		cacheMaxAge:              cacheMaxAge,
	}
}

// @Summary      JSON Web Key Set of partner_app
// @Description  Public keys of every signing key that is pending, active or inactive. my_app polls this to verify activation notification tokens. Cache-Control max-age follows PEER_JWKS_POLL_INTERVAL.
// @Tags         signing-key
// @Produce      json
// @Success      200  {object}  apidoc.JSONWebKeySet
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /.well-known/jwks.json [get]
func (c *SigningKeyController) ListPublishedSigningKeys(ectx echo.Context) error {
	keys, err := c.listPublishedSigningKeys.Execute(ectx.Request().Context())
	if err != nil {
		return err
	}
	ectx.Response().Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", int(c.cacheMaxAge.Seconds())))
	return ectx.JSON(http.StatusOK, c.presenter.ListPublishedSigningKeys(keys))
}

// @Summary      Lifecycle of every signing key (debug)
// @Description  Status is derived from now: pending (published, not signing), active (the only signer), inactive (published, tokens still valid), expired (about to be deleted). Never exposes private key material.
// @Tags         signing-key
// @Produce      json
// @Success      200  {object}  apidoc.SigningKeyList
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /signing-keys [get]
func (c *SigningKeyController) ListSigningKeys(ectx echo.Context) error {
	out, err := c.listSigningKeys.Execute(ectx.Request().Context())
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.ListSigningKeys(out))
}

// @Summary      Rotate the signing key now (debug)
// @Description  Creates the next signing key as pending, published in the JWKS at once and signing after KEY_PENDING_DURATION, and moves the active_until of the current key to that instant so there is no gap and never two signing keys. Rejected while a pending key exists or when no key is active. Not protected: sandbox only.
// @Tags         signing-key
// @Produce      json
// @Success      201  {object}  apidoc.SigningKeyRotation
// @Failure      409  {object}  apidoc.ErrorEnvelope  "SIGNING_KEY_ALREADY_PENDING or SIGNING_KEY_ROTATION_IN_PROGRESS"
// @Failure      503  {object}  apidoc.ErrorEnvelope  "SIGNING_KEY_NOT_FOUND"
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /signing-keys/rotate [post]
func (c *SigningKeyController) RotateSigningKey(ectx echo.Context) error {
	out, err := c.rotateSigningKey.Execute(ectx.Request().Context())
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusCreated, c.presenter.RotateSigningKey(out))
}
