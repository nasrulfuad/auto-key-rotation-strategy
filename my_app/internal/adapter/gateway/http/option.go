package http

import (
	"net/http"
	"time"
)

type Option func(*options)

// options are the dependencies of the HTTP gateways that have a default.
type options struct {
	// httpClient sends the requests; default an http.Client with a 5s timeout.
	httpClient *http.Client
}

func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *options) {
		if httpClient != nil {
			o.httpClient = httpClient
		}
	}
}

func newOptions(opts []Option) options {
	o := options{httpClient: &http.Client{Timeout: 5 * time.Second}}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
