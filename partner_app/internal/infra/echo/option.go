package echo

import (
	"time"

	"github.com/nasrulfuad/jwks_key_rotation/partner_app/internal/infra/logging"
)

type Option func(*options)

// options configure the HTTP server.
type options struct {
	// requestTimeout is the read timeout of the server and the context timeout of every handler; default 15s.
	requestTimeout time.Duration

	// logger receives the request log lines; default logging.NopLogger, which discards them.
	logger logging.Logger
}

func WithRequestTimeout(requestTimeout time.Duration) Option {
	return func(o *options) {
		o.requestTimeout = requestTimeout
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
	o := options{requestTimeout: 15 * time.Second, logger: logging.NopLogger{}}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
