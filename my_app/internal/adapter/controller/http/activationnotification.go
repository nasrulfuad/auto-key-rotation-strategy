package http

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/presenter"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/activationnotification"
)

type ActivationNotificationController struct {
	receiveActivationNotification *activationnotification.ReceiveActivationNotificationInteractor
	listActivationNotifications   *activationnotification.ListActivationNotificationsInteractor
	presenter                     *presenter.ActivationNotificationPresenter
}

func NewActivationNotificationController(
	receiveActivationNotification *activationnotification.ReceiveActivationNotificationInteractor,
	listActivationNotifications *activationnotification.ListActivationNotificationsInteractor,
	activationNotificationPresenter *presenter.ActivationNotificationPresenter,
) *ActivationNotificationController {
	return &ActivationNotificationController{
		receiveActivationNotification: receiveActivationNotification,
		listActivationNotifications:   listActivationNotifications,
		presenter:                     activationNotificationPresenter,
	}
}

type receiveActivationNotificationRequest struct {
	NotificationID string `json:"notification_id"`
	Event          string `json:"event"`
	Provider       string `json:"provider"`
	SubjectID      string `json:"subject_id"`
	EntitlementID  string `json:"entitlement_id"`
	Products       string `json:"products"`
	FreeTrial      string `json:"free_trial"`
}

// @Summary      Activation notification webhook (called by partner_app)
// @Description  Requires a partner_app JWT with aud = ACTIVATION_NOTIFICATION_AUDIENCE; its sub and provider must match subject_id and provider in the body. A repeated notification_id is acknowledged without storing it twice.
// @Tags         activation-notification
// @Accept       json
// @Security     BearerAuth
// @Param        body  body  apidoc.ReceiveActivationNotificationRequest  true  "notification"
// @Success      204
// @Failure      400  {object}  apidoc.ErrorEnvelope  "VALIDATION_FAILED, SUBJECT_ID_INVALID, ENTITLEMENT_ID_INVALID"
// @Failure      401  {object}  apidoc.ErrorEnvelope  "TOKEN_INVALID, TOKEN_ISSUER_UNKNOWN, TOKEN_SIGNING_KEY_NOT_FOUND, TOKEN_CLAIM_INVALID"
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /activation-notifications [post]
func (c *ActivationNotificationController) ReceiveActivationNotification(ectx echo.Context) error {
	token, err := verifiedToken(ectx)
	if err != nil {
		return err
	}
	var request receiveActivationNotificationRequest
	if err := ectx.Bind(&request); err != nil {
		return domain.BadRequest("request body must be valid JSON")
	}
	notificationID, err := uuid.Parse(request.NotificationID)
	if err != nil {
		return domain.Invalid("notification_id", "notification_id must be a UUID")
	}
	if err := c.receiveActivationNotification.Execute(ectx.Request().Context(), activationnotification.ReceiveActivationNotificationInput{
		Token:          token,
		NotificationID: notificationID,
		Event:          request.Event,
		Provider:       request.Provider,
		SubjectID:      request.SubjectID,
		EntitlementID:  request.EntitlementID,
		Products:       request.Products,
		FreeTrial:      request.FreeTrial,
	}); err != nil {
		return err
	}
	return ectx.NoContent(http.StatusNoContent)
}

// @Summary      Latest received activation notifications (debug)
// @Description  Newest first, at most 100. key_id is the partner_app kid that signed each webhook.
// @Tags         activation-notification
// @Produce      json
// @Success      200  {object}  apidoc.ActivationNotificationList
// @Failure      500  {object}  apidoc.ErrorEnvelope
// @Router       /activation-notifications [get]
func (c *ActivationNotificationController) ListActivationNotifications(ectx echo.Context) error {
	notifications, err := c.listActivationNotifications.Execute(ectx.Request().Context())
	if err != nil {
		return err
	}
	return ectx.JSON(http.StatusOK, c.presenter.ListActivationNotifications(notifications))
}
