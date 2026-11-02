//go:build integration

package integration_test

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/di"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/config"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/postgres"
)

var startTime = time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func newDatabase(t *testing.T) *sql.DB {
	t.Helper()
	adminURL := os.Getenv("TEST_DATABASE_ADMIN_URL")
	if adminURL == "" {
		t.Skip("TEST_DATABASE_ADMIN_URL is not set")
	}
	admin, err := sql.Open("pgx", adminURL)
	require.NoError(t, err)
	t.Cleanup(func() { _ = admin.Close() })
	suffix := make([]byte, 6)
	_, err = rand.Read(suffix)
	require.NoError(t, err)
	name := "my_app_test_" + hex.EncodeToString(suffix)
	_, err = admin.Exec("CREATE DATABASE " + name)
	require.NoError(t, err)

	parsed, err := url.Parse(adminURL)
	require.NoError(t, err)
	parsed.Path = "/" + name
	db, err := postgres.Open(parsed.String(), 4)
	require.NoError(t, err)
	require.NoError(t, postgres.Migrate(db))
	t.Cleanup(func() {
		_ = db.Close()
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + name + " WITH (FORCE)")
	})
	return db
}

type app struct {
	container *di.Container
	server    *httptest.Server
	clock     *fakeClock
	config    *config.Config
}

func newApp(t *testing.T, values map[string]string) *app {
	t.Helper()
	settings := map[string]string{
		"DATABASE_URL":                "unused",
		"KEY_ENCRYPTION_KEY":          base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
		"KEY_PENDING_DURATION":        "5m",
		"KEY_ACTIVE_DURATION":         "20m",
		"KEY_INACTIVE_DURATION":       "10m",
		"KEY_ROTATION_CHECK_INTERVAL": "15s",
		"TOKEN_TTL":                   "10m",
		"PEER_JWKS_URL":               "http://127.0.0.1:1/jwks.json",
	}
	for k, v := range values {
		settings[k] = v
	}
	cfg, err := config.FromLookup(func(name string) (string, bool) {
		v, ok := settings[name]
		return v, ok
	})
	require.NoError(t, err)
	clock := &fakeClock{now: startTime}
	container, err := di.NewWithDB(cfg, newDatabase(t), di.WithClock(clock))
	require.NoError(t, err)
	server := httptest.NewServer(container.Echo)
	t.Cleanup(server.Close)
	return &app{container: container, server: server, clock: clock, config: cfg}
}

func (a *app) tick(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	require.NoError(t, a.container.Scheduler.RunOnce(ctx, di.JobEnsureSigningKeys))
	require.NoError(t, a.container.Scheduler.RunOnce(ctx, di.JobDeleteExpiredSigningKeys))
}

type signingKeyView struct {
	KeyID       string    `json:"key_id"`
	Status      string    `json:"status"`
	ActiveFrom  time.Time `json:"active_from"`
	ActiveUntil time.Time `json:"active_until"`
	DeleteAfter time.Time `json:"delete_after"`
}

func (a *app) signingKeys(t *testing.T) []signingKeyView {
	t.Helper()
	response, err := http.Get(a.server.URL + "/signing-keys")
	require.NoError(t, err)
	defer func() { _ = response.Body.Close() }()
	var body struct {
		Data []signingKeyView `json:"data"`
	}
	require.NoError(t, json.NewDecoder(response.Body).Decode(&body))
	return body.Data
}

func (a *app) jwksURL() string { return a.server.URL + "/.well-known/jwks.json" }
