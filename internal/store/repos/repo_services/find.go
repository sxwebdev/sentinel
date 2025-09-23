package repo_services

import (
	"context"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

func findServicesBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From("services s")

	if params.Name != "" {
		sb.Where(sb.Like("s.name", "%"+params.Name+"%"))
	}

	if params.Protocol != "" {
		sb.Where(sb.Equal("s.protocol", params.Protocol))
	}

	if params.IsEnabled != nil {
		sb.Where(sb.Equal("s.is_enabled", *params.IsEnabled))
	}

	if params.Status != "" {
		switch params.Status {
		case "up":
			sb.Where(sb.Equal("ss.status", models.StatusUp))
		case "down":
			sb.Where(sb.Equal("ss.status", models.StatusDown))
		}
	}

	if len(params.Tags) > 0 {
		var tagConditions []string
		for _, tag := range params.Tags {
			tagConditions = append(tagConditions,
				fmt.Sprintf("EXISTS (SELECT 1 FROM json_each(s.tags) WHERE json_each.value = %s)",
					sb.Args.Add(tag)))
		}

		if len(tagConditions) > 0 {
			sb.Where(fmt.Sprintf("(%s)", strings.Join(tagConditions, " AND ")))
		}
	}

	return sb
}

type FindParams struct {
	Name      string
	IsEnabled *bool
	Protocol  string
	Tags      []string
	Status    string // e.g. "up", "down"
	OrderBy   string
	Page      *uint32
	PageSize  *uint32
}

// FindView returns list of services with their states and incidents by given filters with pagination
func (s *CustomQueries) FindView(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.ServiceFullView], error) {
	sb := findServicesBuilder(
		params,
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
		"sum(case when incidents.id IS NOT NULL AND incidents.resolved_at IS NULL then 1 else 0 end) as active_incidents",
		"ss.status",
		"ss.last_check",
		"ss.last_error",
		"ss.consecutive_fails",
		"ss.consecutive_success",
		"ss.total_checks",
		"ss.avg_response_time",
	)
	sb.JoinWithOption(sqlbuilder.LeftJoin, "incidents", "s.id = incidents.service_id")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "service_states ss", "s.id = ss.service_id")
	sb.GroupBy("s.id")

	if params.OrderBy != "" {
		// Add table prefix for common column names to avoid ambiguity
		orderBy := params.OrderBy
		switch orderBy {
		case "created_at":
			orderBy = "s.created_at"
		case "updated_at":
			orderBy = "s.updated_at"
		case "name":
			orderBy = "s.name"
		case "protocol":
			orderBy = "s.protocol"
		case "status":
			orderBy = "ss.status"
		case "last_check":
			orderBy = "ss.last_check"
		}
		sb.OrderBy(orderBy)
	} else {
		sb.OrderBy("s.name")
	}

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	itemsRows := []itemViewRow{}
	sql, args := sb.Build()
	if err := sqlscan.Select(ctx, s.db, &itemsRows, sql, args...); err != nil {
		return nil, err
	}

	// Get total count of services
	countQuery := findServicesBuilder(params, "count(*)")
	countQuery.JoinWithOption(sqlbuilder.LeftJoin, "service_states ss", "s.id = ss.service_id")

	var totalCount uint32
	countSQL, countArgs := countQuery.Build()
	if err := sqlscan.Get(ctx, s.db, &totalCount, countSQL, countArgs...); err != nil {
		return nil, err
	}

	items := make([]*models.ServiceFullView, 0, len(itemsRows))
	for i := range itemsRows {
		item, err := rowToModel(&itemsRows[i])
		if err != nil {
			return nil, fmt.Errorf("failed to convert row to service: %w", err)
		}
		items = append(items, item)
	}

	return &storecmn.FindResponseWithCount[*models.ServiceFullView]{
		Count: totalCount,
		Items: items,
	}, nil
}
