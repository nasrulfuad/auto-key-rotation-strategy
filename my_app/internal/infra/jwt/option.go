package jwt

import (
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/logging"
)

type Option func(*options)

// options are the dependencies of the issuer and the verifier that have a default.
type options struct {
	// logger receives the log lines; default logging.NopLogger, which discards them.
	logger logging.Logger
}

func WithLogger(logger logging.Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

func newOptions(opts []Option) options {
	o := options{logger: logging.NopLogger{}}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
