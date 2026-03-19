package apiserver

import (
	"sync"
	"time"
)

// loginRateLimiter tracks failed login attempts per key (email or IP).
type loginRateLimiter struct {
	mu       sync.Mutex
	attempts map[string][]time.Time
	limit    int
	window   time.Duration
}

func newLoginRateLimiter(limit int, window time.Duration) *loginRateLimiter {
	return &loginRateLimiter{
		attempts: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// isAllowed checks if a new attempt is allowed for the given key.
func (rl *loginRateLimiter) isAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Clean up old attempts
	attempts := rl.attempts[key]
	valid := attempts[:0]
	for _, t := range attempts {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	rl.attempts[key] = valid

	return len(valid) < rl.limit
}

// recordFailure records a failed attempt for the given key.
func (rl *loginRateLimiter) recordFailure(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.attempts[key] = append(rl.attempts[key], time.Now())
}
