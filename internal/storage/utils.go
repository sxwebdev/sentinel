package storage

import (
	"time"
)

// durationToNS converts a duration pointer to nanoseconds pointer
func durationToNS(duration *time.Duration) *int64 {
	if duration == nil {
		return nil
	}
	ns := duration.Nanoseconds()
	return &ns
}
