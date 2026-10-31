package config

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/domain"
)

// Config is the whole runtime configuration of my_app, read from environment variables by FromEnv.
type Config struct {
	// Port is the HTTP port the server listens on. APP_PORT, default 8081.
	Port int

	// DatabaseURL is the PostgreSQL connection string. DATABASE_URL, required.
	DatabaseURL string

	// DatabaseMaxConnections caps the open and idle connections of the pool. DATABASE_MAX_CONNECTIONS, default 5.
	DatabaseMaxConnections int

	// LogLevel is the minimum level written: debug, info, warn or error. LOG_LEVEL, default info.
	LogLevel string

	// RequestTimeout bounds reading a request and handling it. REQUEST_TIMEOUT, default 15s.
	RequestTimeout time.Duration

	// EnableAPIDocs serves Swagger UI at /docs and the spec at /openapi.json. ENABLE_API_DOCS, default false.
	EnableAPIDocs bool

	// TokenIssuer is the iss claim of every token this app signs. TOKEN_ISSUER, default qa.my-app.local.
	TokenIssuer string

	// TokenTTL is the lifetime of every token this app signs; it must not exceed KEY_INACTIVE_DURATION.
	// TOKEN_TTL, default 5m.
	TokenTTL time.Duration

	// TokenLeeway is the clock skew tolerated when checking exp and iat of peer tokens. TOKEN_LEEWAY, default 30s.
	TokenLeeway time.Duration

	// KeyAlgorithm signs this app's tokens: RS256, RS384 or RS512. KEY_ALGORITHM, default RS256.
	KeyAlgorithm domain.SigningAlgorithm

	// KeyRSABits is the modulus size of new signing keys, at least 2048. KEY_RSA_BITS, default 2048.
	KeyRSABits int

	// KeySchedule is the pending, active and inactive duration of every signing key.
	// KEY_PENDING_DURATION, KEY_ACTIVE_DURATION and KEY_INACTIVE_DURATION, default 5m, 20m and 10m.
	KeySchedule domain.SigningKeySchedule

	// KeyRotationCheckInterval is how often the ensure and delete signing key jobs run; it must be shorter than
	// KEY_PENDING_DURATION. KEY_ROTATION_CHECK_INTERVAL, default 15s.
	KeyRotationCheckInterval time.Duration

	// KeyEncryptionKey is the AES-256 key that seals private keys at rest. KEY_ENCRYPTION_KEY, required, base64 of 32 bytes.
	KeyEncryptionKey []byte

	// PeerIssuer is the iss claim expected on partner_app tokens. PEER_ISSUER, default qa.partner-app.local.
	PeerIssuer string

	// PeerJWKSURL is where the JWKS of partner_app is fetched. PEER_JWKS_URL, default http://partner_app:8082/.well-known/jwks.json.
	PeerJWKSURL string

	// PeerKeyAlgorithm is the only algorithm accepted on partner_app tokens. PEER_KEY_ALGORITHM, default RS256.
	PeerKeyAlgorithm domain.SigningAlgorithm

	// PeerJWKSPollInterval is how often the partner_app JWKS is refreshed; a value not shorter than
	// KEY_PENDING_DURATION adds a warning. PEER_JWKS_POLL_INTERVAL, default 1m.
	PeerJWKSPollInterval time.Duration

	// PeerJWKSRefreshOnUnknownKeyID refreshes the partner_app JWKS at once when a token carries an unknown kid.
	// PEER_JWKS_REFRESH_ON_UNKNOWN_KEY_ID, default true.
	PeerJWKSRefreshOnUnknownKeyID bool

	// PeerJWKSRefreshMinInterval is the minimum time between two refreshes triggered by an unknown kid.
	// PEER_JWKS_REFRESH_MIN_INTERVAL, default 10s.
	PeerJWKSRefreshMinInterval time.Duration

	// PartnerBaseURL is the base URL of the partner_app Entitlement API. PARTNER_BASE_URL, default http://partner_app:8082.
	PartnerBaseURL string

	// Provider is the provider claim this app puts on entitlement tokens. PROVIDER, default my_app.
	Provider string

	// EntitlementSetAudience is the aud of tokens sent with entitlement activations and terminations.
	// ENTITLEMENT_SET_AUDIENCE, default set.entitlement.partner-app.local.
	EntitlementSetAudience domain.TokenAudience

	// ActivationNotificationAudience is the aud required on partner_app activation notification webhooks.
	// ACTIVATION_NOTIFICATION_AUDIENCE, default activation-notification.my-app.local.
	ActivationNotificationAudience domain.TokenAudience

	// Warnings lists configuration that is valid but risky; main logs each one at startup.
	Warnings []string
}

func FromEnv() (*Config, error) {
	return FromLookup(os.LookupEnv)
}

func FromLookup(lookup func(string) (string, bool)) (*Config, error) {
	r := reader{lookup: lookup}
	cfg := &Config{
		Port:                   r.integer("APP_PORT", 8081),
		DatabaseURL:            r.required("DATABASE_URL"),
		DatabaseMaxConnections: r.integer("DATABASE_MAX_CONNECTIONS", 5),
		LogLevel:               r.text("LOG_LEVEL", "info"),
		RequestTimeout:         r.duration("REQUEST_TIMEOUT", 15*time.Second),
		EnableAPIDocs:          r.boolean("ENABLE_API_DOCS", false),

		TokenIssuer: r.text("TOKEN_ISSUER", "qa.my-app.local"),
		TokenTTL:    r.duration("TOKEN_TTL", 5*time.Minute),
		TokenLeeway: r.duration("TOKEN_LEEWAY", 30*time.Second),

		KeyAlgorithm: domain.SigningAlgorithm(r.text("KEY_ALGORITHM", string(domain.SigningAlgorithmRS256))),
		KeyRSABits:   r.integer("KEY_RSA_BITS", 2048),
		KeySchedule: domain.SigningKeySchedule{
			PendingDuration:  r.duration("KEY_PENDING_DURATION", 5*time.Minute),
			ActiveDuration:   r.duration("KEY_ACTIVE_DURATION", 20*time.Minute),
			InactiveDuration: r.duration("KEY_INACTIVE_DURATION", 10*time.Minute),
		},
		KeyRotationCheckInterval: r.duration("KEY_ROTATION_CHECK_INTERVAL", 15*time.Second),
		KeyEncryptionKey:         r.base64Key("KEY_ENCRYPTION_KEY"),

		PeerIssuer:                    r.text("PEER_ISSUER", "qa.partner-app.local"),
		PeerJWKSURL:                   r.text("PEER_JWKS_URL", "http://partner_app:8082/.well-known/jwks.json"),
		PeerKeyAlgorithm:              domain.SigningAlgorithm(r.text("PEER_KEY_ALGORITHM", string(domain.SigningAlgorithmRS256))),
		PeerJWKSPollInterval:          r.duration("PEER_JWKS_POLL_INTERVAL", time.Minute),
		PeerJWKSRefreshOnUnknownKeyID: r.boolean("PEER_JWKS_REFRESH_ON_UNKNOWN_KEY_ID", true),
		PeerJWKSRefreshMinInterval:    r.duration("PEER_JWKS_REFRESH_MIN_INTERVAL", 10*time.Second),

		PartnerBaseURL:                 r.text("PARTNER_BASE_URL", "http://partner_app:8082"),
		Provider:                       r.text("PROVIDER", "my_app"),
		EntitlementSetAudience:         domain.TokenAudience(r.text("ENTITLEMENT_SET_AUDIENCE", "set.entitlement.partner-app.local")),
		ActivationNotificationAudience: domain.TokenAudience(r.text("ACTIVATION_NOTIFICATION_AUDIENCE", "activation-notification.my-app.local")),
	}
	if len(r.errs) > 0 {
		return nil, errors.Join(r.errs...)
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	var errs []error
	if err := c.KeySchedule.Validate(c.TokenTTL); err != nil {
		errs = append(errs, err)
	}
	if !c.KeyAlgorithm.Valid() {
		errs = append(errs, fmt.Errorf("KEY_ALGORITHM must be RS256, RS384 or RS512, got %q", c.KeyAlgorithm))
	}
	if !c.PeerKeyAlgorithm.Valid() {
		errs = append(errs, fmt.Errorf("PEER_KEY_ALGORITHM must be RS256, RS384 or RS512, got %q", c.PeerKeyAlgorithm))
	}
	if c.KeyRSABits < 2048 {
		errs = append(errs, fmt.Errorf("KEY_RSA_BITS must be at least 2048, got %d", c.KeyRSABits))
	}
	if c.KeyRotationCheckInterval <= 0 || c.KeyRotationCheckInterval >= c.KeySchedule.PendingDuration {
		errs = append(errs, fmt.Errorf("KEY_ROTATION_CHECK_INTERVAL %s must be positive and shorter than KEY_PENDING_DURATION %s",
			c.KeyRotationCheckInterval, c.KeySchedule.PendingDuration))
	}
	if c.PeerJWKSPollInterval <= 0 {
		errs = append(errs, fmt.Errorf("PEER_JWKS_POLL_INTERVAL must be positive, got %s", c.PeerJWKSPollInterval))
	}
	if c.TokenLeeway < 0 {
		errs = append(errs, fmt.Errorf("TOKEN_LEEWAY must not be negative, got %s", c.TokenLeeway))
	}
	if c.PeerJWKSPollInterval >= c.KeySchedule.PendingDuration {
		c.Warnings = append(c.Warnings, fmt.Sprintf(
			"PEER_JWKS_POLL_INTERVAL %s is not shorter than KEY_PENDING_DURATION %s; a peer with the same schedule may sign with a key this app has not polled yet",
			c.PeerJWKSPollInterval, c.KeySchedule.PendingDuration))
	}
	return errors.Join(errs...)
}

type reader struct {
	lookup func(string) (string, bool)
	errs   []error
}

func (r *reader) value(name string) (string, bool) {
	value, ok := r.lookup(name)
	value = strings.TrimSpace(value)
	return value, ok && value != ""
}

func (r *reader) required(name string) string {
	value, ok := r.value(name)
	if !ok {
		r.errs = append(r.errs, fmt.Errorf("%s is required", name))
	}
	return value
}

func (r *reader) text(name, fallback string) string {
	if value, ok := r.value(name); ok {
		return value
	}
	return fallback
}

func (r *reader) integer(name string, fallback int) int {
	value, ok := r.value(name)
	if !ok {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s must be an integer: %w", name, err))
		return fallback
	}
	return parsed
}

func (r *reader) duration(name string, fallback time.Duration) time.Duration {
	value, ok := r.value(name)
	if !ok {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s must be a duration such as 30s or 5m: %w", name, err))
		return fallback
	}
	return parsed
}

func (r *reader) boolean(name string, fallback bool) bool {
	value, ok := r.value(name)
	if !ok {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s must be true or false: %w", name, err))
		return fallback
	}
	return parsed
}

func (r *reader) base64Key(name string) []byte {
	value := r.required(name)
	if value == "" {
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		r.errs = append(r.errs, fmt.Errorf("%s must be base64: %w", name, err))
		return nil
	}
	if len(decoded) != 32 {
		r.errs = append(r.errs, fmt.Errorf("%s must decode to 32 bytes, got %d", name, len(decoded)))
		return nil
	}
	return decoded
}
