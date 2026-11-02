package http

import (
	"strings"

	"github.com/labstack/echo/v4"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/domain"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/usecase/authentication"
)

const verifiedTokenKey = "verified_token"

type PeerAuthMiddleware struct {
	authenticatePeerRequest *authentication.AuthenticatePeerRequestInteractor
	logger                  logging.Logger
}

func NewPeerAuthMiddleware(authenticatePeerRequest *authentication.AuthenticatePeerRequestInteractor, opts ...Option) *PeerAuthMiddleware {
	return &PeerAuthMiddleware{authenticatePeerRequest: authenticatePeerRequest, logger: newOptions(opts).logger}
}

func (m *PeerAuthMiddleware) Require(audience domain.TokenAudience) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ectx echo.Context) error {
			ctx := ectx.Request().Context()
			token, err := m.authenticatePeerRequest.Execute(ctx, authentication.AuthenticatePeerRequestInput{
				RawToken: bearerToken(ectx.Request().Header.Get(echo.HeaderAuthorization)),
				Audience: audience,
			})
			if err != nil {
				return err
			}
			m.logger.Info(ctx, "peer request authenticated",
				"kid", token.KeyID, "iss", token.Claims.Issuer, "aud", audience,
				"provider", token.Claims.Provider, "jti", token.Claims.ID, "route", ectx.Path())
			ectx.Set(verifiedTokenKey, token)
			return next(ectx)
		}
	}
}

func bearerToken(header string) string {
	scheme, token, found := strings.Cut(header, " ")
	if !found || !strings.EqualFold(scheme, "Bearer") {
		return ""
	}
	return strings.TrimSpace(token)
}

func verifiedToken(ectx echo.Context) (domain.VerifiedToken, error) {
	token, ok := ectx.Get(verifiedTokenKey).(domain.VerifiedToken)
	if !ok {
		return domain.VerifiedToken{}, domain.Internalf("route %s has no verified token", ectx.Path())
	}
	return token, nil
}
