// @title           partner_app API
// @version         1.0
// @description     Provider side of the JWKS key rotation simulation: Entitlement API protected by my_app JWTs, its own rotating JWKS, and debug views of both key sets. Every failure uses one error envelope; see apidoc.ErrorEnvelope.
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 "Bearer <JWT signed by my_app>". The token must carry kid, iss = PEER_ISSUER, provider = PEER_PROVIDER and the aud of the route.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	partnerapp "github.com/nasrulfuad/auto-key-rotation-strategy/partner_app"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/di"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/config"
	echoserver "github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/echo"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/postgres"
)

const (
	serviceName     = "partner_app"
	shutdownTimeout = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("startup failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	_ = godotenv.Load()

	cfg, err := config.FromEnv()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	logger := logging.NewJSON(os.Stdout, logging.WithLevel(cfg.LogLevel), logging.WithService(serviceName))
	startupCtx := context.Background()
	for _, warning := range cfg.Warnings {
		logger.Warn(startupCtx, warning)
	}

	openAPISpec, err := partnerapp.OpenAPIFS.ReadFile("api/openapi.json")
	if err != nil {
		return fmt.Errorf("failed to read the embedded OpenAPI spec: %w", err)
	}

	container, err := di.New(cfg, di.WithLogger(logger), di.WithOpenAPISpec(openAPISpec))
	if err != nil {
		return err
	}
	defer func() {
		if err := container.Close(); err != nil {
			logger.Error(startupCtx, "failed to close dependency", "error", err)
		}
	}()

	if err := postgres.Migrate(container.DB); err != nil {
		return fmt.Errorf("failed to run database migrations: %w", err)
	}
	logger.Info(startupCtx, "migrations applied")

	logger.Info(startupCtx, "signing key schedule",
		"algorithm", cfg.KeyAlgorithm,
		"pending", cfg.KeySchedule.PendingDuration.String(),
		"active", cfg.KeySchedule.ActiveDuration.String(),
		"inactive", cfg.KeySchedule.InactiveDuration.String(),
		"token_ttl", cfg.TokenTTL.String(),
		"issuer", cfg.TokenIssuer,
		"peer_issuer", cfg.PeerIssuer,
		"peer_jwks_url", cfg.PeerJWKSURL)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := container.Scheduler.RunOnce(ctx, di.JobEnsureSigningKeys); err != nil {
		return fmt.Errorf("failed to ensure a signing key at startup: %w", err)
	}
	container.StartBackground(ctx)

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.Port)
	server := echoserver.NewServer(container.Echo, addr, echoserver.WithRequestTimeout(cfg.RequestTimeout))
	serverErr := make(chan error, 1)
	go func() {
		logger.Info(ctx, serviceName+" listening", "addr", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("server error: %w", err)
		}
	case <-ctx.Done():
		logger.Info(ctx, "shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	logger.Info(shutdownCtx, "shutdown complete")
	return nil
}
