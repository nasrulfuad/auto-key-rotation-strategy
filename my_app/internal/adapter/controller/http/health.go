package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/presenter"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/usecase/port"
)

type HealthController struct {
	databaseProbe port.DatabaseProbe
	presenter     *presenter.HealthPresenter
}

func NewHealthController(databaseProbe port.DatabaseProbe, healthPresenter *presenter.HealthPresenter) *HealthController {
	return &HealthController{databaseProbe: databaseProbe, presenter: healthPresenter}
}

// @Summary  Liveness probe
// @Tags     health
// @Produce  json
// @Success  200  {object}  apidoc.Health
// @Router   /health [get]
func (c *HealthController) CheckLiveness(ectx echo.Context) error {
	return ectx.JSON(http.StatusOK, c.presenter.CheckLiveness())
}

// @Summary  Readiness probe (database round-trip)
// @Tags     health
// @Produce  json
// @Success  200  {object}  apidoc.Health
// @Failure  503  {object}  apidoc.ErrorEnvelope  "DATABASE_UNREACHABLE"
// @Router   /health/db [get]
func (c *HealthController) CheckReadiness(ectx echo.Context) error {
	if err := c.databaseProbe.Ping(ectx.Request().Context()); err != nil {
		return domain.ServiceUnavailableCoded("DATABASE_UNREACHABLE", "database unreachable")
	}
	return ectx.JSON(http.StatusOK, c.presenter.CheckReadiness())
}
