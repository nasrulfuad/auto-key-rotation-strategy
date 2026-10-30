package systemclock_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/jwks_key_rotation/my_app/internal/infra/systemclock"
)

func TestSystemNowReturnsCurrentTimeInUTC(t *testing.T) {
	before := time.Now()

	now := systemclock.New().Now()

	assert.Equal(t, time.UTC, now.Location())
	assert.False(t, now.Before(before.Add(-time.Second)))
	assert.False(t, now.After(time.Now().Add(time.Second)))
}

func TestUUIDv4NewUUIDReturnsDistinctVersion4UUIDs(t *testing.T) {
	generator := systemclock.NewUUIDGenerator()

	first, second := generator.NewUUID(), generator.NewUUID()

	assert.Equal(t, uuid.Version(4), first.Version())
	assert.NotEqual(t, first, second)
}
