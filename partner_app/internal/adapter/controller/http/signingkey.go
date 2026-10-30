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
	presenter                *presenter.SigningKeyPresenter

	cacheMaxAge time.Duration
}

func NewSigningKeyController(
	listPublishedSigningKeys *signingkey.ListPublishedSigningKeysInteractor,
	listSigningKeys *signingkey.ListSigningKeysInteractor,
	signingKeyPresenter *presenter.SigningKeyPresenter,
	cacheMaxAge time.Duration,
) *SigningKeyController {
	return &SigningKeyController{
		listPublishedSigningKeys: listPublishedSigningKeys,
		listSigningKeys:          listSigningKeys,
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
