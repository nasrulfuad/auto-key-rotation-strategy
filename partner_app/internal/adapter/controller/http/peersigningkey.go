package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/presenter"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/peersigningkey"
)

type PeerSigningKeyController struct {
	listPeerSigningKeys *peersigningkey.ListPeerSigningKeysInteractor
	presenter           *presenter.PeerSigningKeyPresenter
}

func NewPeerSigningKeyController(
	listPeerSigningKeys *peersigningkey.ListPeerSigningKeysInteractor,
	peerSigningKeyPresenter *presenter.PeerSigningKeyPresenter,
	opts ...Option,
) *PeerSigningKeyController {
	return &PeerSigningKeyController{listPeerSigningKeys: listPeerSigningKeys, presenter: peerSigningKeyPresenter}
}

// @Summary      Cached JWKS of my_app (debug)
// @Description  Keys this app currently trusts to verify my_app tokens, plus the result of the last fetch.
// @Tags         signing-key
// @Produce      json
// @Success      200  {object}  apidoc.PeerSigningKeyList
// @Router       /peer-signing-keys [get]
func (c *PeerSigningKeyController) ListPeerSigningKeys(ectx *echo.Context) error {
	out := c.listPeerSigningKeys.Execute(ectx.Request().Context())
	return ectx.JSON(http.StatusOK, c.presenter.ListPeerSigningKeys(out))
}
