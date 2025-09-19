package incidents

import (
	"context"
	"time"

	"github.com/sxwebdev/sentinel/internal/store/repos/repo_incidents"
)

// StatsByDateRange retrieves the stats of incidents within a specific date range
func (s *Service) StatsByDateRange(ctx context.Context, startTime, endTime time.Time) (repo_incidents.StatsByDateRangeData, error) {
	return s.store.Incidents().StatsByDateRange(ctx, startTime, endTime)
}
