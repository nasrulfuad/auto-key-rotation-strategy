package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/domain"
)

type Access int

const (
	Public Access = iota
	PeerAuthenticated
)

type Route struct {
	Method   string
	Path     string
	Handler  echo.HandlerFunc
	Access   Access
	Audience domain.TokenAudience
}

type Audiences struct {
	EntitlementSet domain.TokenAudience
	EntitlementGet domain.TokenAudience
}

type Router struct {
	health             *HealthController
	signingKey         *SigningKeyController
	peerSigningKey     *PeerSigningKeyController
	entitlement        *EntitlementController
	peerAuthMiddleware *PeerAuthMiddleware
	docs               *DocsController
	audiences          Audiences
}

func NewRouter(
	health *HealthController,
	signingKey *SigningKeyController,
	peerSigningKey *PeerSigningKeyController,
	entitlement *EntitlementController,
	peerAuthMiddleware *PeerAuthMiddleware,
	docs *DocsController,
	audiences Audiences,
) *Router {
	return &Router{
		health:             health,
		signingKey:         signingKey,
		peerSigningKey:     peerSigningKey,
		entitlement:        entitlement,
		peerAuthMiddleware: peerAuthMiddleware,
		docs:               docs,
		audiences:          audiences,
	}
}

func (r *Router) Table() []Route {
	table := []Route{
		{Method: http.MethodGet, Path: "/health", Handler: r.health.CheckLiveness, Access: Public},
		{Method: http.MethodGet, Path: "/health/db", Handler: r.health.CheckReadiness, Access: Public},
		{Method: http.MethodGet, Path: "/.well-known/jwks.json", Handler: r.signingKey.ListPublishedSigningKeys, Access: Public},
		{Method: http.MethodGet, Path: "/signing-keys", Handler: r.signingKey.ListSigningKeys, Access: Public},
		{Method: http.MethodGet, Path: "/peer-signing-keys", Handler: r.peerSigningKey.ListPeerSigningKeys, Access: Public},
		{Method: http.MethodPut, Path: "/subjects/:subject_id/entitlements/:entitlement_id", Handler: r.entitlement.UpsertEntitlement, Access: PeerAuthenticated, Audience: r.audiences.EntitlementSet},
		{Method: http.MethodGet, Path: "/subjects/:subject_id/entitlements/:entitlement_id", Handler: r.entitlement.GetEntitlement, Access: PeerAuthenticated, Audience: r.audiences.EntitlementGet},
		{Method: http.MethodGet, Path: "/subjects/:subject_id/entitlements", Handler: r.entitlement.ListEntitlements, Access: PeerAuthenticated, Audience: r.audiences.EntitlementGet},
		{Method: http.MethodPost, Path: "/subjects/:subject_id/entitlements/:entitlement_id/terminate", Handler: r.entitlement.TerminateEntitlement, Access: PeerAuthenticated, Audience: r.audiences.EntitlementSet},
	}
	if r.docs != nil {
		table = append(table,
			Route{Method: http.MethodGet, Path: "/openapi.json", Handler: r.docs.OpenAPI, Access: Public},
			Route{Method: http.MethodGet, Path: "/docs", Handler: r.docs.UI, Access: Public},
		)
	}
	return table
}

func (r *Router) Mount(e *echo.Echo, table []Route) {
	for _, route := range table {
		var middlewares []echo.MiddlewareFunc
		switch route.Access {
		case Public:
		case PeerAuthenticated:
			if route.Audience == "" {
				panic(fmt.Sprintf("route %s %s requires peer authentication but has no audience", route.Method, route.Path))
			}
			middlewares = append(middlewares, r.peerAuthMiddleware.Require(route.Audience))
		default:
			panic(fmt.Sprintf("route %s %s has an unknown access level %d", route.Method, route.Path, route.Access))
		}
		e.Add(route.Method, route.Path, route.Handler, middlewares...)
	}
}
