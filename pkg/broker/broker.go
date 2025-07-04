package broker

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Broker - a thread-safe broker for subscribers.
//   - T is the type of messages.
//   - All operations (Subscribe/Unsubscribe/Publish) can be called from different goroutines.
//   - Start() launches an internal goroutine to distribute messages.
//   - Stop() stops its operation and closes all subscriber channels.
type Broker[T any] struct {
	// subs stores all active subscriber channels using sync.Map,
	// to avoid blocking Subscribe/Unsubscribe during subscriber iteration.
	subs sync.Map // key: chan T, value: struct{}

	// publishCh - incoming messages for publishing.
	// Can be buffered if you want to avoid blocking Publish.
	publishCh chan T

	// stopCh - signal to stop the broker.
	stopCh chan struct{}

	// doneCh - signal that the broker's goroutine has completely stopped.
	doneCh chan struct{}

	// closed - a flag indicating that the broker is stopped (to prevent new subscriptions after Stop).
	closed atomic.Bool
}

// NewBroker initializes the broker but does NOT start its goroutine.
func NewBroker[T any]() *Broker[T] {
	return &Broker[T]{
		publishCh: make(chan T, 16), // can increase the buffer size if needed
		stopCh:    make(chan struct{}),
		doneCh:    make(chan struct{}),
	}
}

// Start starts the broker's goroutine: it reads from publishCh and distributes messages.
// Should be called once, usually via go broker.Start().
func (b *Broker[T]) Start() {
	defer close(b.doneCh)

	for {
		select {
		case <-b.stopCh:
			// Close all subscriber channels and exit
			b.closeAllSubscribers()
			return

		case msg, ok := <-b.publishCh:
			if !ok {
				// If publishCh is unexpectedly closed, terminate.
				b.closeAllSubscribers()
				return
			}
			// Broadcast the message to all subscribers
			b.broadcast(msg)
		}
	}
}

// Stop stops the broker: closes stopCh and waits for the goroutine to exit.
func (b *Broker[T]) Stop() {
	// Set the flag to prevent new subscriptions/publications
	if b.closed.CompareAndSwap(false, true) {
		close(b.stopCh) // signal the broker's goroutine
		<-b.doneCh      // wait for the goroutine to fully terminate
		// We don't close publishCh here to avoid a panic in other goroutines
		// if someone continues to call Publish (we show below how to ignore it).
	}
}

// Subscribe returns a new channel for receiving messages.
// If the broker is already closed, it returns nil.
func (b *Broker[T]) Subscribe() chan T {
	if b.closed.Load() {
		return nil
	}
	ch := make(chan T, 8) // buffered channel to avoid blocking the broker
	b.subs.Store(ch, struct{}{})
	return ch
}

// Unsubscribe removes a channel from the list and closes it (if present).
func (b *Broker[T]) Unsubscribe(ch chan T) {
	// Attempt to remove the subscription (if not already removed)
	if _, ok := b.subs.Load(ch); ok {
		b.subs.Delete(ch)
		close(ch)
	}
}

// Publish sends a message to the internal queue.
// If the broker is already closed, it ignores the message or returns.
func (b *Broker[T]) Publish(msg T) {
	if b.closed.Load() {
		// The broker is stopped - do not accept new messages.
		return
	}
	// You can either block on send or use select { case ... default: } as preferred.
	b.publishCh <- msg
}

// --- Helper methods ---

// broadcast sends msg to all subscribers.
// On panic (e.g., if a subscriber closed their channel), it removes the subscription from subs.
func (b *Broker[T]) broadcast(msg T) {
	b.subs.Range(func(key, _ any) bool {
		ch := key.(chan T) //nolint:forcetypeassert
		safeSend(ch, msg, func(ch chan T) {
			// If the send causes a panic (channel is closed),
			// remove it from the list of subscribers.
			b.subs.Delete(ch)
		})
		return true
	})
}

// closeAllSubscribers closes all channels in subs.
func (b *Broker[T]) closeAllSubscribers() {
	b.subs.Range(func(key, _ any) bool {
		ch := key.(chan T) //nolint:forcetypeassert
		b.subs.Delete(ch)
		close(ch)
		return true
	})
}

// safeSend performs a non-blocking send to a channel with panic recovery.
// removeOnPanic is called if the channel is found to be closed (panic on send).
func safeSend[T any](ch chan T, msg T, removeOnPanic func(ch chan T)) {
	defer func() {
		if r := recover(); r != nil {
			// The channel is likely closed by the subscriber
			fmt.Printf("subscriber channel is closed, removing from subs (panic: %v)\n", r) //nolint:forbidigo
			removeOnPanic(ch)
		}
	}()
	select {
	case ch <- msg:
	default:
		// If there's no space in the channel's buffer - "skip" the message.
		// You can remove default if you want to wait for the subscriber to read.
	}
}
