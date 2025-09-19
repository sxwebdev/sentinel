package repo_services

import (
	"context"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
)

func (o *CustomQueries) GetAllTags(ctx context.Context) ([]string, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("DISTINCT json_each.value")
	sb.From("services, json_each(tags)")
	sb.OrderBy("json_each.value")

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tags, nil
}

func (o *CustomQueries) GetAllTagsWithCount(ctx context.Context) (map[string]int, error) {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("json_each.value, COUNT(*)")
	sb.From("services, json_each(tags)")
	sb.GroupBy("json_each.value")
	sb.OrderBy("json_each.value")

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query tags with count: %w", err)
	}
	defer rows.Close()

	tagCounts := make(map[string]int)
	for rows.Next() {
		var tag string
		var count int
		if err := rows.Scan(&tag, &count); err != nil {
			return nil, fmt.Errorf("failed to scan tag with count: %w", err)
		}
		tagCounts[tag] = count
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return tagCounts, nil
}
