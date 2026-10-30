package config_test

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/config"
)

func lookup(values map[string]string) func(string) (string, bool) {
	base := map[string]string{
		"DATABASE_URL":       "postgres://localhost/db",
		"KEY_ENCRYPTION_KEY": base64.StdEncoding.EncodeToString([]byte(strings.Repeat("k", 32))),
	}
	for k, v := range values {
		base[k] = v
	}
	return func(name string) (string, bool) {
		v, ok := base[name]
		return v, ok
	}
}

func TestFromLookupAppliesDefaultSchedule(t *testing.T) {
	cfg, err := config.FromLookup(lookup(nil))

	require.NoError(t, err)
	assert.Equal(t, 5*time.Minute, cfg.KeySchedule.PendingDuration)
	assert.Equal(t, 20*time.Minute, cfg.KeySchedule.ActiveDuration)
	assert.Equal(t, 10*time.Minute, cfg.KeySchedule.InactiveDuration)
	assert.Empty(t, cfg.Warnings)
}

func TestFromLookupRejectsInvalidConfiguration(t *testing.T) {
	cases := map[string]map[string]string{
		"missing database":          {"DATABASE_URL": ""},
		"token outlives inactive":   {"TOKEN_TTL": "11m"},
		"short encryption key":      {"KEY_ENCRYPTION_KEY": base64.StdEncoding.EncodeToString([]byte("short"))},
		"weak RSA key":              {"KEY_RSA_BITS": "1024"},
		"unsupported algorithm":     {"KEY_ALGORITHM": "HS256"},
		"check slower than pending": {"KEY_ROTATION_CHECK_INTERVAL": "5m"},
		"unparsable duration":       {"KEY_ACTIVE_DURATION": "twenty minutes"},
		"zero inactive duration":    {"KEY_INACTIVE_DURATION": "0s", "TOKEN_TTL": "1s"},
	}
	for name, values := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := config.FromLookup(lookup(values))
			assert.Error(t, err)
		})
	}
}

func TestFromLookupWarnsWhenPollIsSlowerThanPendingPhase(t *testing.T) {
	cfg, err := config.FromLookup(lookup(map[string]string{"PEER_JWKS_POLL_INTERVAL": "10m"}))

	require.NoError(t, err)
	assert.Len(t, cfg.Warnings, 1)
}
