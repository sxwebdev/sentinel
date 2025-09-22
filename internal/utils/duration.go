package utils

import (
	"fmt"
	"time"
)

// FormatDuration formats a duration in a human-readable way
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}

	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		seconds := int(d.Seconds()) % 60
		if seconds > 0 {
			return fmt.Sprintf("%dm %ds", minutes, seconds)
		}
		return fmt.Sprintf("%dm", minutes)
	}

	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	// For durations >= 24 hours
	totalHours := int(d.Hours())
	days := totalHours / 24
	hours := totalHours % 24
	minutes := int(d.Minutes()) % 60

	if minutes > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}

	if hours > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}

	return fmt.Sprintf("%dd", days)
}
