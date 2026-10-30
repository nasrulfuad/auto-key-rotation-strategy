// @title           my_app API
// @version         1.0
// @description     Partner side of the JWKS key rotation simulation: demo endpoints that call partner_app with JWTs signed by the active key, the activation notification webhook, its own rotating JWKS, and debug views of both key sets. Every failure uses one error envelope; see apidoc.ErrorEnvelope.
// @securityDefinitions.apikey  BearerAuth
// @in                          header
// @name                        Authorization
// @description                 "Bearer <JWT signed by partner_app>". The token must carry kid, iss = PEER_ISSUER and aud = ACTIVATION_NOTIFICATION_AUDIENCE.
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

	myapp "github.com/nasrulfuad/jwks_key_rotation/my_app"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/di"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/config"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/logging"
	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/postgres"
)

const (
	serviceName     = "my_app"
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
	logger := logging.NewJSON(os.Stdout, cfg.LogLevel, serviceName)
	startupCtx := context.Background()
	for _, warning := range cfg.Warnings {
		logger.Warn(startupCtx, warning)
	}

	openAPISpec, err := myapp.OpenAPIFS.ReadFile("api/openapi.json")
	if err != nil {
		return fmt.Errorf("failed to read the embedded OpenAPI spec: %w", err)
	}

	container, err := di.New(cfg, logger, openAPISpec)
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
	serverErr := make(chan error, 1)
	go func() {
		logger.Info(ctx, serviceName+" listening", "addr", addr)
		if err := container.Echo.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
	if err := container.Echo.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("graceful shutdown failed: %w", err)
	}
	logger.Info(shutdownCtx, "shutdown complete")
	return nil
}
