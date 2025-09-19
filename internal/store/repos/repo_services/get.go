package repo_services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

// GetViewByID finds service by ID
func (s *CustomQueries) GetViewByID(ctx context.Context, id string) (*models.ServiceFullView, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"s.id",
		"s.name",
		"s.protocol",
		"s.interval",
		"s.timeout",
		"s.retries",
		"s.tags",
		"s.config",
		"s.is_enabled",
		"s.created_at",
		"s.updated_at",
		"count(incidents.id) as total_incidents",
		"sum(case when incidents.resolved = 0 then 1 else 0 end) as active_incidents",
		"ss.status",
		"ss.last_check",
		"ss.next_check",
		"ss.last_error",
		"ss.consecutive_fails",
		"ss.consecutive_success",
		"ss.total_checks",
		"ss.response_time",
	)
	sb.From("services s")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "incidents", "s.id = incidents.service_id")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "service_states ss", "s.id = ss.service_id")
	sb.Where(sb.Equal("s.id", id))
	sb.GroupBy("s.id")

	var itemRow itemViewRow
	query, args := sb.Build()
	if err := sqlscan.Get(ctx, s.db, &itemRow, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storecmn.ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan service: %w", err)
	}

	item, err := rowToModel(&itemRow)
	if err != nil {
		return nil, fmt.Errorf("failed to convert row to service: %w", err)
	}

	return item, nil
}
