package activationnotification

import (
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/systemclock"
)

type Option func(*options)

// options are the dependencies of the activationnotification use cases that have a default.
type options struct {
	// clock tells the current time; default systemclock.New().
	clock systemclock.Clock

	// uuidGenerator generates row identifiers and token IDs; default systemclock.NewUUIDGenerator().
	uuidGenerator systemclock.UUIDGenerator

	// logger receives the use case log lines; default logging.NopLogger, which discards them.
	logger logging.Logger
}

func WithClock(clock systemclock.Clock) Option {
	return func(o *options) {
		if clock != nil {
			o.clock = clock
		}
	}
}

func WithUUIDGenerator(uuidGenerator systemclock.UUIDGenerator) Option {
	return func(o *options) {
		if uuidGenerator != nil {
			o.uuidGenerator = uuidGenerator
		}
	}
}

func WithLogger(logger logging.Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

func newOptions(opts []Option) options {
	o := options{clock: systemclock.New(), uuidGenerator: systemclock.NewUUIDGenerator(), logger: logging.NopLogger{}}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
