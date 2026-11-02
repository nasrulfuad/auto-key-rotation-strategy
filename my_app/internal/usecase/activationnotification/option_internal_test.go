package activationnotification

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/systemclock"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time { return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC) }

type fixedUUIDGenerator struct{}

func (fixedUUIDGenerator) NewUUID() uuid.UUID { return uuid.Nil }

func TestNewOptionsDefaultsToSystemClockUUIDv4AndNopLogger(t *testing.T) {
	o := newOptions(nil)

	assert.Equal(t, systemclock.New(), o.clock)
	assert.Equal(t, systemclock.NewUUIDGenerator(), o.uuidGenerator)
	assert.Equal(t, logging.NopLogger{}, o.logger)
}

func TestNewOptionsAppliesEveryOption(t *testing.T) {
	logger := logging.NewJSON(nil)

	o := newOptions([]Option{WithClock(fixedClock{}), WithUUIDGenerator(fixedUUIDGenerator{}), WithLogger(logger)})

	assert.Equal(t, fixedClock{}, o.clock)
	assert.Equal(t, fixedUUIDGenerator{}, o.uuidGenerator)
	assert.Equal(t, logger, o.logger)
}

func TestNewOptionsIgnoresNilDependencies(t *testing.T) {
	o := newOptions([]Option{WithClock(nil), WithUUIDGenerator(nil), WithLogger(nil)})

	assert.Equal(t, systemclock.New(), o.clock)
	assert.Equal(t, systemclock.NewUUIDGenerator(), o.uuidGenerator)
	assert.Equal(t, logging.NopLogger{}, o.logger)
}
