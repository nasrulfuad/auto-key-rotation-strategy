package rsakey

type Option func(*options)

// options configure the key pair generator.
type options struct {
	// bits is the modulus size of new keys; default and minimum MinimumBits (2048).
	bits int
}

func WithBits(bits int) Option {
	return func(o *options) {
		o.bits = bits
	}
}

func newOptions(opts []Option) options {
	o := options{bits: MinimumBits}
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
