package echo

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	controller "github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/controller/http"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/adapter/presenter"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/logging"
)

const (
	maximumRequestBody = 1 << 20
	readHeaderTimeout  = 5 * time.Second
	idleTimeout        = 2 * time.Minute
	HeaderLogID        = "X-Log-ID"
)

func New(router *controller.Router, errorPresenter *presenter.ErrorPresenter, opts ...Option) *echo.Echo {
	o := newOptions(opts)
	logger := o.logger
	e := echo.New()
	e.HTTPErrorHandler = errorPresenter.Render

	e.Use(logRequest(logger))
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(maximumRequestBody))
	e.Use(middleware.ContextTimeout(o.requestTimeout))

	router.Mount(e, router.Table())
	return e
}

func NewServer(handler http.Handler, address string, opts ...Option) *http.Server {
	o := newOptions(opts)
	return &http.Server{
		Addr:              address,
		Handler:           handler,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       o.requestTimeout,
		IdleTimeout:       idleTimeout,
	}
}

func logRequest(logger logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ectx *echo.Context) error {
			startedAt := time.Now()
			logID := uuid.NewString()
			ectx.Response().Header().Set(HeaderLogID, logID)
			ectx.SetRequest(ectx.Request().WithContext(logging.AddContextFields(ectx.Request().Context(),
				logging.FieldSource, logging.SourceAPI, logging.FieldLogID, logID)))
			err := next(ectx)
			if err != nil {
				ectx.Echo().HTTPErrorHandler(ectx, err)
			}
			request := ectx.Request()
			logger.Info(context.WithoutCancel(request.Context()), "request completed",
				"method", request.Method, "uri", request.RequestURI, "path", request.URL.Path, "route", ectx.Path(),
				"status", responseStatus(ectx), "duration_ms", time.Since(startedAt).Milliseconds())
			return nil
		}
	}
}

func responseStatus(ectx *echo.Context) int {
	response, err := echo.UnwrapResponse(ectx.Response())
	if err != nil {
		return 0
	}
	return response.Status
}
