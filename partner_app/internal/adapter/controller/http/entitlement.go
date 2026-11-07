package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/presenter"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/entitlement"
)

type EntitlementController struct {
	upsertEntitlement    *entitlement.UpsertEntitlementInteractor
	getEntitlement       *entitlement.GetEntitlementInteractor
	listEntitlements     *entitlement.ListEntitlementsInteractor
	terminateEntitlement *entitlement.TerminateEntitlementInteractor
	presenter            *presenter.EntitlementPresenter
}

func NewEntitlementController(
	upsertEntitlement *entitlement.UpsertEntitlementInteractor,
	getEntitlement *entitlement.GetEntitlementInteractor,
	listEntitlements *entitlement.ListEntitlementsInteractor,
	terminateEntitlement *entitlement.TerminateEntitlementInteractor,
	entitlementPresenter *presenter.EntitlementPresenter,
	opts ...Option,
) *EntitlementController {
	return &EntitlementController{
		upsertEntitlement:    upsertEntitlement,
		getEntitlement:       getEntitlement,
		listEntitlements:     listEntitlements,
		terminateEntitlement: terminateEntitlement,
		presenter:            entitlementPresenter,
	}
}

type upsertEntitlementRequest struct {
	Products string       `json:"products"`
	Offer    offerRequest `json:"offer"`
}

type offerRequest struct {
	FreeTrial string `json:"free_trial"`
}

// @Summary      put_entitlement: activate (create, update or reactivate) one entitlement
// @Description  Requires a my_app JWT with aud = ENTITLEMENT_SET_AUDIENCE (default set.entitlement.partner-app.local) and provider = PEER_PROVIDER. Sets status ACTIVE and clears terminated_at. The token iat becomes last_updated_at; an iat older than the stored value is rejected. Queues an entitlement_upserted activation notification to my_app.
// @Tags         entitlement
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        subject_id      path  string                           true  "customer identifier, [0-9a-zA-Z-._=]+"  example(SUB001)
// @Param        entitlement_id  path  string                           true  "entitlement identifier, [0-9a-zA-Z-._=]+"  example(ENT001)
// @Param        body            body  apidoc.UpsertEntitlementRequest  true  "entitlement"
// @Success      200  {object}  apidoc.Entitlement
// @Failure      400  {object}  apidoc.ErrorEnvelope  "VALIDATION_FAILED, SUBJECT_ID_INVALID, ENTITLEMENT_ID_INVALID, ENTITLEMENT_STALE_UPDATE"
// @Failure      401  {object}  apidoc.ErrorEnvelope  "TOKEN_INVALID, TOKEN_ISSUER_UNKNOWN, TOKEN_SIGNING_KEY_NOT_FOUND, TOKEN_CLAIM_INVALID"
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /subjects/{subject_id}/entitlements/{entitlement_id} [put]
func (c *EntitlementController) UpsertEntitlement(ectx *echo.Context) error {
	token, err := verifiedToken(ectx)
	if err != nil {
		return err
	}
	var request upsertEntitlementRequest
	if err := ectx.Bind(&request); err != nil {
		return domain.BadRequest("request body must be valid JSON")
	}
	result, err := c.upsertEntitlement.Execute(ectx.Request().Context(), entitlement.UpsertEntitlementInput{
		Token:         token,
		SubjectID:     ectx.Param("subject_id"),
		EntitlementID: ectx.Param("entitlement_id"),
		Products:      request.Products,
		FreeTrial:     request.Offer.FreeTrial,
	})
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.GetEntitlement(result))
}

// @Summary      get_entitlement: read one entitlement
// @Description  Requires a my_app JWT with aud = ENTITLEMENT_GET_AUDIENCE (default get.entitlement.partner-app.local). Scoped to the token provider.
// @Tags         entitlement
// @Produce      json
// @Security     BearerAuth
// @Param        subject_id      path  string  true  "customer identifier"  example(SUB001)
// @Param        entitlement_id  path  string  true  "entitlement identifier"  example(ENT001)
// @Success      200  {object}  apidoc.Entitlement
// @Failure      400  {object}  apidoc.ErrorEnvelope  "SUBJECT_ID_INVALID, ENTITLEMENT_ID_INVALID"
// @Failure      401  {object}  apidoc.ErrorEnvelope  "TOKEN_INVALID, TOKEN_ISSUER_UNKNOWN, TOKEN_SIGNING_KEY_NOT_FOUND, TOKEN_CLAIM_INVALID"
// @Failure      404  {object}  apidoc.ErrorEnvelope  "ENTITLEMENT_NOT_FOUND"
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /subjects/{subject_id}/entitlements/{entitlement_id} [get]
func (c *EntitlementController) GetEntitlement(ectx *echo.Context) error {
	token, err := verifiedToken(ectx)
	if err != nil {
		return err
	}
	result, err := c.getEntitlement.Execute(ectx.Request().Context(), entitlement.GetEntitlementInput{
		Token:         token,
		SubjectID:     ectx.Param("subject_id"),
		EntitlementID: ectx.Param("entitlement_id"),
	})
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.GetEntitlement(result))
}

// @Summary      List every entitlement of one subject
// @Description  Requires a my_app JWT with aud = ENTITLEMENT_GET_AUDIENCE. Scoped to the token provider.
// @Tags         entitlement
// @Produce      json
// @Security     BearerAuth
// @Param        subject_id  path  string  true  "customer identifier"  example(SUB001)
// @Success      200  {object}  apidoc.EntitlementList
// @Failure      400  {object}  apidoc.ErrorEnvelope  "SUBJECT_ID_INVALID"
// @Failure      401  {object}  apidoc.ErrorEnvelope  "TOKEN_INVALID, TOKEN_ISSUER_UNKNOWN, TOKEN_SIGNING_KEY_NOT_FOUND, TOKEN_CLAIM_INVALID"
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /subjects/{subject_id}/entitlements [get]
func (c *EntitlementController) ListEntitlements(ectx *echo.Context) error {
	token, err := verifiedToken(ectx)
	if err != nil {
		return err
	}
	result, err := c.listEntitlements.Execute(ectx.Request().Context(), entitlement.ListEntitlementsInput{
		Token:     token,
		SubjectID: ectx.Param("subject_id"),
	})
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.ListEntitlements(result))
}

// @Summary      Terminate one entitlement
// @Description  Requires a my_app JWT with aud = ENTITLEMENT_SET_AUDIENCE and provider = PEER_PROVIDER. Sets status TERMINATED and terminated_at (kept on repeated calls). The token iat becomes last_updated_at; an iat older than the stored value is rejected. Queues an entitlement_terminated activation notification to my_app.
// @Tags         entitlement
// @Produce      json
// @Security     BearerAuth
// @Param        subject_id      path  string  true  "customer identifier"  example(SUB001)
// @Param        entitlement_id  path  string  true  "entitlement identifier"  example(ENT001)
// @Success      200  {object}  apidoc.Entitlement
// @Failure      400  {object}  apidoc.ErrorEnvelope  "SUBJECT_ID_INVALID, ENTITLEMENT_ID_INVALID, ENTITLEMENT_STALE_UPDATE"
// @Failure      401  {object}  apidoc.ErrorEnvelope  "TOKEN_INVALID, TOKEN_ISSUER_UNKNOWN, TOKEN_SIGNING_KEY_NOT_FOUND, TOKEN_CLAIM_INVALID"
// @Failure      404  {object}  apidoc.ErrorEnvelope  "ENTITLEMENT_NOT_FOUND"
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /subjects/{subject_id}/entitlements/{entitlement_id}/terminate [post]
func (c *EntitlementController) TerminateEntitlement(ectx *echo.Context) error {
	token, err := verifiedToken(ectx)
	if err != nil {
		return err
	}
	result, err := c.terminateEntitlement.Execute(ectx.Request().Context(), entitlement.TerminateEntitlementInput{
		Token:         token,
		SubjectID:     ectx.Param("subject_id"),
		EntitlementID: ectx.Param("entitlement_id"),
	})
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.GetEntitlement(result))
}
