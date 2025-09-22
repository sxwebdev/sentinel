package loop

import (
	"context"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type Loop struct {
	options Options

	running atomic.Bool
	stopped atomic.Bool
	wg      sync.WaitGroup
	cancel  context.CancelFunc

	fn func(context.Context)
}

func New(fn func(context.Context), opts ...Option) *Loop {
	if fn == nil {
		panic("function cannot be nil")
	}

	options := Options{
		period:         time.Second * 60,
		contextTimeout: time.Second * 30,
		logger:         &emptyLogger{},
	}

	for _, o := range opts {
		o(&options)
	}

	return &Loop{
		options: options,
		fn:      fn,
	}
}

func (l *Loop) Start(parent context.Context) {
	ctx, cancel := context.WithCancel(parent)
	l.cancel = cancel

	// Run immediately if leading is true
	if l.options.leading {
		l.tryRun(ctx, l.fn)
	}

	l.wg.Go(func() {
		t := time.NewTicker(l.options.period)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if l.stopped.Load() {
					return
				}
				l.tryRun(ctx, l.fn)
			}
		}
	})
}

// Trigger triggers the loop to run the provided function immediately if not already running
func (l *Loop) Trigger(ctx context.Context) bool {
	return l.tryRun(ctx, l.fn)
}

// Stop stops the loop
func (l *Loop) Stop() {
	if l.stopped.Swap(true) {
		return
	}
	if l.cancel != nil {
		l.cancel()
	}
}

// Wait waits for all running operations to complete (graceful shutdown)
func (l *Loop) Wait() {
	l.wg.Wait()
}

// tryRun tries to run a function within the loop's context
func (l *Loop) tryRun(parent context.Context, fn func(context.Context)) bool {
	if l.stopped.Load() {
		return false
	}
	if !l.running.CompareAndSwap(false, true) {
		return false
	}

	l.wg.Go(func() {
		now := time.Now()
		defer func() {
			if r := recover(); r != nil {
				l.options.logger.Errorf("panic recovered in loop: %v\n%s", r, debug.Stack())
			}
			l.running.Store(false)
			l.options.logger.Debugf("loop iteration took %s", time.Since(now))
		}()

		ctx := parent
		var cancel context.CancelFunc
		if l.options.contextTimeout > 0 {
			ctx, cancel = context.WithTimeout(parent, l.options.contextTimeout)
			defer cancel()
		}

		fn(ctx)
	})

	return true
}
