package repo_incidents

import (
	"context"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

type StatsByDateRangeItem struct {
	Date          time.Time `json:"date"`
	Count         int64     `json:"count"`
	AvgDuration   int64     `json:"avg_duration"`
	TotalDuration int64     `json:"total_duration"`
}

type StatsByDateRangeData []StatsByDateRangeItem

// StatsByDateRange retrieves the stats of incidents within a specific date range
func (o *CustomQueries) StatsByDateRange(ctx context.Context, startTime, endTime time.Time) (StatsByDateRangeData, error) {
	// Generate date series for the range
	var result StatsByDateRangeData

	// Iterate through each day in the range
	for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
		// Get start and end of the day
		dayStart := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		dayEnd := time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 999999999, d.Location())

		// Query incidents for this specific date range
		sb := sqlbuilder.NewSelectBuilder()
		sb.Select(
			"COUNT(*) as count",
			"AVG(CASE WHEN resolved = true THEN duration ELSE (strftime('%s', 'now') - strftime('%s', start_time)) * 1000 END) as avg_duration",
			"SUM(CASE WHEN resolved = true THEN duration ELSE (strftime('%s', 'now') - strftime('%s', start_time)) * 1000 END) as total_duration",
		)
		sb.From("incidents")
		sb.Where(sb.GreaterEqualThan("start_time", dayStart))
		sb.Where(sb.LessEqualThan("start_time", dayEnd))

		sql, args := sb.Build()
		row := o.db.QueryRowContext(ctx, sql, args...)

		var count int64
		var avgDuration *float64
		var totalDuration *float64

		if err := row.Scan(&count, &avgDuration, &totalDuration); err != nil {
			return nil, fmt.Errorf("failed to scan incident stats for date %s: %w", d.Format("2006-01-02"), err)
		}

		item := StatsByDateRangeItem{
			Date:  d,
			Count: count,
		}

		// Convert nanoseconds to duration
		if avgDuration != nil {
			item.AvgDuration = int64(*avgDuration)
		}

		if totalDuration != nil {
			item.TotalDuration = int64(*totalDuration)
		}

		result = append(result, item)
	}

	return result, nil
}
