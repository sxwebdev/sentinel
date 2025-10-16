package repo_agents

import (
	"context"
	"fmt"
	"strings"

	"github.com/georgysavva/scany/v2/sqlscan"
	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type FindParams struct {
	Name      string
	IsEnabled *bool
	Tags      []string
	Status    string
	ProjectID string
	OrderBy   string
	Page      *uint32
	PageSize  *uint32
}

func findBuilder(params FindParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From(TableNameAgents.String()).
		Where(sb.Equal(ColumnNameAgentsProjectId.String(), params.ProjectID))

	if params.Name != "" {
		sb.Where(sb.Like("name", "%"+params.Name+"%"))
	}

	if params.IsEnabled != nil {
		sb.Where(sb.Equal("is_enabled", *params.IsEnabled))
	}

	if params.Status != "" {
		sb.Where(sb.Equal("status", params.Status))
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

// Find returns list of agents by given filters with pagination
func (s *CustomQueries) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Agent], error) {
	sb := findBuilder(params, AgentsColumnNames().Strings()...)

	if params.OrderBy == "" {
		params.OrderBy = "created_at"
	}

	sb.OrderByDesc(params.OrderBy)

	limit, offset, err := storecmn.Pagination(params.Page, params.PageSize)
	if err != nil {
		return nil, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	items := []*models.Agent{}
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

	return &storecmn.FindResponseWithCount[*models.Agent]{
		Count: totalCount,
		Items: items,
	}, nil
}
