package repos

import "database/sql"

type Option func(*Options)

type Options struct {
	Tx *sql.Tx
}

func parseOptions(opts ...Option) Options {
	options := Options{}
	for _, opt := range opts {
		opt(&options)
	}

	return options
}

func WithTx(tx *sql.Tx) Option {
	return func(o *Options) {
		o.Tx = tx
	}
}
