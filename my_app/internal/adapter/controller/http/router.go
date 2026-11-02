package http

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/domain"
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

// Audiences are the aud values the peer authentication middleware requires per route.
type Audiences struct {
	// ActivationNotification is required on POST /activation-notifications.
	ActivationNotification domain.TokenAudience
}

type Router struct {
	health                 *HealthController
	signingKey             *SigningKeyController
	peerSigningKey         *PeerSigningKeyController
	subscription           *SubscriptionController
	activationNotification *ActivationNotificationController
	peerAuthMiddleware     *PeerAuthMiddleware
	docs                   *DocsController
	audiences              Audiences
}

func NewRouter(
	health *HealthController,
	signingKey *SigningKeyController,
	peerSigningKey *PeerSigningKeyController,
	subscription *SubscriptionController,
	activationNotification *ActivationNotificationController,
	peerAuthMiddleware *PeerAuthMiddleware,
	audiences Audiences,
	opts ...Option,
) *Router {
	return &Router{
		health:                 health,
		signingKey:             signingKey,
		peerSigningKey:         peerSigningKey,
		subscription:           subscription,
		activationNotification: activationNotification,
		peerAuthMiddleware:     peerAuthMiddleware,
		docs:                   newOptions(opts).docs,
		audiences:              audiences,
	}
}

func (r *Router) Table() []Route {
	table := []Route{
		{Method: http.MethodGet, Path: "/health", Handler: r.health.CheckLiveness, Access: Public},
		{Method: http.MethodGet, Path: "/health/db", Handler: r.health.CheckReadiness, Access: Public},
		{Method: http.MethodGet, Path: "/.well-known/jwks.json", Handler: r.signingKey.ListPublishedSigningKeys, Access: Public},
		{Method: http.MethodGet, Path: "/signing-keys", Handler: r.signingKey.ListSigningKeys, Access: Public},
		{Method: http.MethodPost, Path: "/signing-keys/rotate", Handler: r.signingKey.RotateSigningKey, Access: Public},
		{Method: http.MethodGet, Path: "/peer-signing-keys", Handler: r.peerSigningKey.ListPeerSigningKeys, Access: Public},
		{Method: http.MethodPost, Path: "/activations", Handler: r.subscription.ActivateSubscription, Access: Public},
		{Method: http.MethodPost, Path: "/terminations", Handler: r.subscription.TerminateSubscription, Access: Public},
		{Method: http.MethodGet, Path: "/subscriptions", Handler: r.subscription.ListSubscriptions, Access: Public},
		{Method: http.MethodPost, Path: "/activation-notifications", Handler: r.activationNotification.ReceiveActivationNotification, Access: PeerAuthenticated, Audience: r.audiences.ActivationNotification},
		{Method: http.MethodGet, Path: "/activation-notifications", Handler: r.activationNotification.ListActivationNotifications, Access: Public},
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
