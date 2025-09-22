package loop

import "time"

type Options struct {
	logger         logger
	leading        bool
	period         time.Duration
	contextTimeout time.Duration
}

type Option func(*Options)

// WithLogger sets the logger for the loop.
func WithLogger(logger logger) Option {
	return func(o *Options) {
		if logger == nil {
			return
		}
		o.logger = logger
	}
}

// WithLeading sets the loop to execute the task immediately upon starting.
func WithLeading() Option {
	return func(o *Options) {
		o.leading = true
	}
}

// WithPeriod sets the period for the loop execution.
func WithPeriod(d time.Duration) Option {
	return func(o *Options) {
		if d <= 0 {
			return
		}
		o.period = d
	}
}

// WithContextTimeout sets a timeout for the context passed to the task function.
func WithContextTimeout(d time.Duration) Option {
	return func(o *Options) {
		o.contextTimeout = d
	}
}
