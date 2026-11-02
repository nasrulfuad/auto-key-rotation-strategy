package di

import (
	"net/http"
	"time"

	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/my_app/internal/infra/systemclock"
)

type Option func(*options)

// options replace dependencies of the container, mainly for tests.
type options struct {
	// logger receives every log line of the app; default logging.NopLogger, which discards them.
	logger logging.Logger

	// clock tells the current time to every use case and the JWKS cache; default systemclock.New().
	clock systemclock.Clock

	// httpClient is used for every outbound call; default an http.Client with a 5s timeout.
	httpClient *http.Client

	// openAPISpec is served at /openapi.json when ENABLE_API_DOCS is true; default empty.
	openAPISpec []byte
}

func WithLogger(logger logging.Logger) Option {
	return func(o *options) {
		if logger != nil {
			o.logger = logger
		}
	}
}

func WithClock(clock systemclock.Clock) Option {
	return func(o *options) {
		if clock != nil {
			o.clock = clock
		}
	}
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *options) {
		if httpClient != nil {
			o.httpClient = httpClient
		}
	}
}

func WithOpenAPISpec(openAPISpec []byte) Option {
	return func(o *options) {
		if openAPISpec != nil {
			o.openAPISpec = openAPISpec
		}
	}
}

func newOptions(opts []Option) options {
	o := options{logger: logging.NopLogger{}, clock: systemclock.New(), httpClient: &http.Client{Timeout: 5 * time.Second}}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
