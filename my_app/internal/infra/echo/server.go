package echo

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	controller "github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/controller/http"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/adapter/presenter"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
)

const (
	maximumRequestBody = "1M"
	readHeaderTimeout  = 5 * time.Second
	idleTimeout        = 2 * time.Minute
	HeaderLogID        = "X-Log-ID"
)

// Options configure the HTTP server.
type Options struct {
	// RequestTimeout is the read timeout of the server and the context timeout of every handler.
	RequestTimeout time.Duration

	// Logger receives the request log lines; nil discards them.
	Logger logging.Logger
}

func New(o Options, router *controller.Router, errorPresenter *presenter.ErrorPresenter) *echo.Echo {
	logger := o.Logger
	if logger == nil {
		logger = logging.NopLogger{}
	}
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.HTTPErrorHandler = errorPresenter.Render
	e.Server.ReadHeaderTimeout = readHeaderTimeout
	e.Server.ReadTimeout = o.RequestTimeout
	e.Server.IdleTimeout = idleTimeout

	e.Use(logRequest(logger))
	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(maximumRequestBody))
	e.Use(middleware.ContextTimeout(o.RequestTimeout))

	router.Mount(e, router.Table())
	return e
}

func logRequest(logger logging.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ectx echo.Context) error {
			startedAt := time.Now()
			logID := uuid.NewString()
			ectx.Response().Header().Set(HeaderLogID, logID)
			ectx.SetRequest(ectx.Request().WithContext(logging.AddContextFields(ectx.Request().Context(),
				logging.FieldSource, logging.SourceAPI, logging.FieldLogID, logID)))
			err := next(ectx)
			if err != nil {
				ectx.Error(err)
			}
			request := ectx.Request()
			logger.Info(context.WithoutCancel(request.Context()), "request completed",
				"method", request.Method, "uri", request.RequestURI, "path", request.URL.Path, "route", ectx.Path(),
				"status", ectx.Response().Status, "duration_ms", time.Since(startedAt).Milliseconds())
			return nil
		}
	}
}
