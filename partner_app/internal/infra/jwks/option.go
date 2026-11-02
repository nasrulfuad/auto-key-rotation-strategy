package jwks

import (
	"net/http"
	"time"

	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/logging"
	"github.com/nasrulfuad/auto-key-rotation-strategy/partner_app/internal/infra/systemclock"
)

type Option func(*options)

// options configure a RemoteKeySet.
type options struct {
	// httpClient fetches the JWKS; default an http.Client with a 5s timeout.
	httpClient *http.Client

	// refreshOnUnknownKeyID makes FindPublicKey refresh the key set when keyID is not cached; default true.
	refreshOnUnknownKeyID bool

	// refreshMinimumInterval is the minimum time between two refreshes triggered by an unknown key ID, default 10s;
	// after a failed refresh the next one is allowed after at most one second.
	refreshMinimumInterval time.Duration

	// clock tells the current time; default systemclock.New().
	clock systemclock.Clock

	// logger receives the log lines; default logging.NopLogger, which discards them.
	logger logging.Logger
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *options) {
		if httpClient != nil {
			o.httpClient = httpClient
		}
	}
}

func WithRefreshOnUnknownKeyID(refreshOnUnknownKeyID bool) Option {
	return func(o *options) {
		o.refreshOnUnknownKeyID = refreshOnUnknownKeyID
	}
}

func WithRefreshMinimumInterval(refreshMinimumInterval time.Duration) Option {
	return func(o *options) {
		o.refreshMinimumInterval = refreshMinimumInterval
	}
}

func WithClock(clock systemclock.Clock) Option {
	return func(o *options) {
		if clock != nil {
			o.clock = clock
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
	o := options{httpClient: &http.Client{Timeout: 5 * time.Second}, refreshOnUnknownKeyID: true, refreshMinimumInterval: 10 * time.Second, clock: systemclock.New(), logger: logging.NopLogger{}}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
