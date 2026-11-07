package http

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/adapter/presenter"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/usecase/subscription"
)

type SubscriptionController struct {
	activateSubscription  *subscription.ActivateSubscriptionInteractor
	terminateSubscription *subscription.TerminateSubscriptionInteractor
	listSubscriptions     *subscription.ListSubscriptionsInteractor
	presenter             *presenter.SubscriptionPresenter
}

func NewSubscriptionController(
	activateSubscription *subscription.ActivateSubscriptionInteractor,
	terminateSubscription *subscription.TerminateSubscriptionInteractor,
	listSubscriptions *subscription.ListSubscriptionsInteractor,
	subscriptionPresenter *presenter.SubscriptionPresenter,
	opts ...Option,
) *SubscriptionController {
	return &SubscriptionController{
		activateSubscription:  activateSubscription,
		terminateSubscription: terminateSubscription,
		listSubscriptions:     listSubscriptions,
		presenter:             subscriptionPresenter,
	}
}

type activateSubscriptionRequest struct {
	SubjectID     string       `json:"subject_id"`
	EntitlementID string       `json:"entitlement_id"`
	Products      string       `json:"products"`
	Offer         offerRequest `json:"offer"`
}

type offerRequest struct {
	FreeTrial string `json:"free_trial"`
}

type terminateSubscriptionRequest struct {
	SubjectID     string `json:"subject_id"`
	EntitlementID string `json:"entitlement_id"`
}

// @Summary      Activate a subscription at partner_app
// @Description  Signs a JWT with the currently active my_app key (aud = ENTITLEMENT_SET_AUDIENCE, provider = PROVIDER), calls PUT /subjects/{subject_id}/entitlements/{entitlement_id} on partner_app, then stores the result locally. signed_with_key_id shows which kid signed the call, so a rotation is visible from here. Activating a terminated subscription reactivates it.
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        body  body  apidoc.ActivateSubscriptionRequest  true  "activation"
// @Success      200  {object}  apidoc.Subscription
// @Failure      400  {object}  apidoc.ErrorEnvelope  "VALIDATION_FAILED, SUBJECT_ID_INVALID, ENTITLEMENT_ID_INVALID"
// @Failure      502  {object}  apidoc.ErrorEnvelope  "PARTNER_REQUEST_REJECTED (message carries partner_app status and code), PARTNER_UNAVAILABLE"
// @Failure      503  {object}  apidoc.ErrorEnvelope  "SIGNING_KEY_NOT_FOUND"
// @Router       /activations [post]
func (c *SubscriptionController) ActivateSubscription(ectx *echo.Context) error {
	var request activateSubscriptionRequest
	if err := ectx.Bind(&request); err != nil {
		return domain.BadRequest("request body must be valid JSON")
	}
	result, err := c.activateSubscription.Execute(ectx.Request().Context(), subscription.ActivateSubscriptionInput{
		SubjectID:     request.SubjectID,
		EntitlementID: request.EntitlementID,
		Products:      request.Products,
		FreeTrial:     request.Offer.FreeTrial,
	})
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.GetSubscription(result))
}

// @Summary      Terminate a subscription at partner_app
// @Description  Signs a JWT with aud = ENTITLEMENT_SET_AUDIENCE, calls POST /subjects/{subject_id}/entitlements/{entitlement_id}/terminate on partner_app, then stores the result locally.
// @Tags         subscription
// @Accept       json
// @Produce      json
// @Param        body  body  apidoc.TerminateSubscriptionRequest  true  "termination"
// @Success      200  {object}  apidoc.Subscription
// @Failure      400  {object}  apidoc.ErrorEnvelope  "SUBJECT_ID_INVALID, ENTITLEMENT_ID_INVALID"
// @Failure      502  {object}  apidoc.ErrorEnvelope  "PARTNER_REQUEST_REJECTED (404 ENTITLEMENT_NOT_FOUND when it was never activated), PARTNER_UNAVAILABLE"
// @Failure      503  {object}  apidoc.ErrorEnvelope  "SIGNING_KEY_NOT_FOUND"
// @Router       /terminations [post]
func (c *SubscriptionController) TerminateSubscription(ectx *echo.Context) error {
	var request terminateSubscriptionRequest
	if err := ectx.Bind(&request); err != nil {
		return domain.BadRequest("request body must be valid JSON")
	}
	result, err := c.terminateSubscription.Execute(ectx.Request().Context(), subscription.TerminateSubscriptionInput{
		SubjectID:     request.SubjectID,
		EntitlementID: request.EntitlementID,
	})
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.GetSubscription(result))
}

// @Summary      Latest local subscriptions (debug)
// @Description  Newest first, at most 100. signed_with_key_id is the my_app kid that signed the last call to partner_app.
// @Tags         subscription
// @Produce      json
// @Success      200  {object}  apidoc.SubscriptionList
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /subscriptions [get]
func (c *SubscriptionController) ListSubscriptions(ectx *echo.Context) error {
	subscriptions, err := c.listSubscriptions.Execute(ectx.Request().Context())
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.ListSubscriptions(subscriptions))
}
