package presenter

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
)

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    string  `json:"code"`
	Field   *string `json:"field,omitempty"`
	Message string  `json:"message"`
}

func Status(kind domain.Kind) int {
	switch kind {
	case domain.KindNotFound:
		return http.StatusNotFound
	case domain.KindUnauthorized:
		return http.StatusUnauthorized
	case domain.KindBadRequest:
		return http.StatusBadRequest
	case domain.KindServiceUnavailable:
		return http.StatusServiceUnavailable
	case domain.KindBadGateway:
		return http.StatusBadGateway
	case domain.KindConflict:
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func Code(err *domain.Error) string {
	if code := err.Code(); code != "" {
		return code
	}
	switch err.Kind() {
	case domain.KindDB:
		return "DB_ERROR"
	case domain.KindNotFound:
		return "NOT_FOUND"
	case domain.KindUnauthorized:
		return "UNAUTHORIZED"
	case domain.KindBadRequest:
		return "BAD_REQUEST"
	case domain.KindServiceUnavailable:
		return "SERVICE_UNAVAILABLE"
	case domain.KindBadGateway:
		return "BAD_GATEWAY"
	case domain.KindConflict:
		return "CONFLICT"
	default:
		return "INTERNAL_ERROR"
	}
}

func PublicMessage(err *domain.Error) string {
	switch err.Kind() {
	case domain.KindNotFound:
		return "resource not found"
	case domain.KindUnauthorized:
		return "authentication required"
	case domain.KindBadRequest, domain.KindServiceUnavailable, domain.KindBadGateway, domain.KindConflict:
		return err.Message()
	default:
		return "internal server error"
	}
}

type ErrorPresenter struct {
	logger logging.Logger
}

func NewErrorPresenter(opts ...Option) *ErrorPresenter {
	return &ErrorPresenter{logger: newOptions(opts).logger}
}

func (p *ErrorPresenter) Render(err error, ectx echo.Context) {
	appErr := toDomainError(err)
	status := Status(appErr.Kind())
	request := ectx.Request()
	ctx := request.Context()
	fields := []any{
		"error", appErr.Error(), "response_code", Code(appErr), "status", status,
		"method", request.Method, "route", ectx.Path(), "path", request.URL.Path,
	}
	if status >= http.StatusInternalServerError {
		p.logger.Error(ctx, "request failed", fields...)
	} else {
		p.logger.Warn(ctx, "request rejected", fields...)
	}
	if ectx.Response().Committed {
		return
	}
	var field *string
	if f := appErr.Field(); f != "" {
		field = &f
	}
	if writeErr := ectx.JSON(status, errorBody{Error: errorDetail{
		Code:    Code(appErr),
		Field:   field,
		Message: PublicMessage(appErr),
	}}); writeErr != nil {
		p.logger.Error(ctx, "failed to write error envelope", "error", writeErr)
	}
}

func toDomainError(err error) *domain.Error {
	var appErr *domain.Error
	if errors.As(err, &appErr) {
		return appErr
	}
	var httpErr *echo.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.Code {
		case http.StatusNotFound:
			return domain.NotFound("route_not_found")
		case http.StatusBadRequest, http.StatusUnsupportedMediaType, http.StatusMethodNotAllowed, http.StatusRequestEntityTooLarge:
			return domain.BadRequest(httpErrorMessage(httpErr))
		case http.StatusServiceUnavailable:
			return domain.ServiceUnavailableCoded("SERVICE_UNAVAILABLE", "service unavailable")
		default:
			return domain.Internal(err)
		}
	}
	return domain.Internal(err)
}

func httpErrorMessage(httpErr *echo.HTTPError) string {
	if msg, ok := httpErr.Message.(string); ok && msg != "" {
		return msg
	}
	return http.StatusText(httpErr.Code)
}
