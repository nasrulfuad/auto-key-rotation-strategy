package http

import (
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/logging"
)

type Option func(*options)

// options are the dependencies of the controllers and the router that have a default.
type options struct {
	// logger receives the log lines; default logging.NopLogger, which discards them.
	logger logging.Logger

	// docs serves /docs and /openapi.json; default nil, which leaves both routes out.
	docs *DocsController
}

func WithLogger(logger logging.Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

func WithDocs(docs *DocsController) Option {
	return func(o *options) {
		if docs != nil {
			o.docs = docs
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
