package repo_notification_history

import (
	"context"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type FindParams struct {
	Status   string
	OrderBy  string
	Page     *uint32
	PageSize *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameNotificationHistory.String() + " h")

	if params.Status != "" {
		sb.Where(sb.Equal("status", params.Status))
	}

	return sb
}

// Find returns list of notification histories by given filters with pagination
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.NotificationHistoryView], error) {
	columns := NotificationHistoryColumnNames().Strings()
	for i := range columns {
		columns[i] = "h." + columns[i]
	}
	columns = append(columns, "s.name AS service_name")

	sb := findBuilder(params, columns...)
	sb.JoinWithOption(sqlbuilder.LeftJoin, "services s", "h.service_id is not null and h.service_id = s.id")

	if params.OrderBy != "" {
		sb.OrderBy(params.OrderBy)
	} else {
		sb.OrderBy("h.created_at")
	}

	sb.Desc()

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.NotificationHistoryView{}
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

	return &storecmn.FindResponseWithCount[*models.NotificationHistoryView]{
		Count: totalCount,
		Items: items,
	}, nil
}
