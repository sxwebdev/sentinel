package repo_incidents

import (
	"context"
	"fmt"
	"time"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type FindParams struct {
	Search    string
	ID        string
	ServiceID string
	Resolved  *bool
	StartTime *time.Time
	EndTime   *time.Time
	Page      *uint32
	PageSize  *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameIncidents.String())

	if params.ID != "" {
		sb.Where(sb.Equal("id", params.ID))
	}

	if params.ServiceID != "" {
		sb.Where(sb.Equal("service_id", params.ServiceID))
	}

	if params.Search != "" {
		likeCondition := fmt.Sprintf("%%%s%%", params.Search)
		sb.Where(sb.Or(
			sb.Like("id", likeCondition),
			sb.Like("service_id", likeCondition),
		))
	}

	if params.Resolved != nil {
		sb.Where(sb.Equal("resolved", *params.Resolved))
	}

	if params.StartTime != nil {
		sb.Where(sb.GreaterEqualThan("start_time", *params.StartTime))
	}

	if params.EndTime != nil {
		sb.Where(sb.LessEqualThan("end_time", *params.EndTime))
	}

	return sb
}

// Find returns list of incidents by given filters with pagination
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Incident], error) {
	sb := findBuilder(params, IncidentsColumnNames().Strings()...)

	sb.OrderBy("created_at").Desc()

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.Incident{}
	sql, args := sb.Build()
	if err := sqlscan.Select(ctx, s.db, &items, sql, args...); err != nil {
		return nil, err
	}

	// Get total count of services
	var totalCount uint32
	countSQL, countArgs := findBuilder(params, "count(*)").Build()
	if err := sqlscan.Get(ctx, s.db, &totalCount, countSQL, countArgs...); err != nil {
		return nil, err
	}

	return &storecmn.FindResponseWithCount[*models.Incident]{
		Count: totalCount,
		Items: items,
	}, nil
}
