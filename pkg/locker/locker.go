package locker

import "sync"

type Locker[T any] struct {
	mu   *sync.RWMutex
	item T
}

func New[T any](item T) *Locker[T] {
	return &Locker[T]{
		mu:   &sync.RWMutex{},
		item: item,
	}
}

// Get returns the locked item
func (l *Locker[T]) Get() T {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.item
}

// Set sets the locked item
func (l *Locker[T]) Set(item T) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.item = item
}
