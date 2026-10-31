package logging

import (
	"log/slog"
)

type Option func(*options)

// options configure a SlogLogger.
type options struct {
	// handler writes the records of New; default the handler of slog.Default().
	handler slog.Handler

	// level is the minimum level NewJSON writes: debug, info, warn or error; default info.
	level string

	// service is the value of the service attribute NewJSON adds to every line; default empty.
	service string
}

func WithHandler(handler slog.Handler) Option {
	return func(o *options) {
		if handler != nil {
			o.handler = handler
		}
	}
}

func WithLevel(level string) Option {
	return func(o *options) {
		o.level = level
	}
}

func WithService(service string) Option {
	return func(o *options) {
		o.service = service
	}
}

func newOptions(opts []Option) options {
	o := options{handler: slog.Default().Handler(), level: "info"}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
