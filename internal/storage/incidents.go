package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

// IncidentRow represents a database row for incidents
type IncidentRow struct {
	ID         string     `db:"id"`
	ServiceID  string     `db:"service_id"`
	StartTime  time.Time  `db:"start_time"`
	EndTime    *time.Time `db:"end_time"`
	Error      string     `db:"error"`
	DurationNS *int64     `db:"duration_ns"`
	Resolved   bool       `db:"resolved"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

// GetIncidentByID retrieves an incident by ID
func (o *ORMStorage) GetIncidentByID(ctx context.Context, id string) (*Incident, error) {
	if id == "" {
		return nil, fmt.Errorf("id is required")
	}

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"i.id",
		"i.service_id",
		"i.start_time",
		"i.end_time",
		"i.error",
		"i.duration_ns",
		"i.resolved",
		"i.created_at",
		"i.updated_at",
	)
	sb.From("incidents i")
	sb.Where(sb.Equal("i.id", id))

	query, args := sb.Build()
	row := o.db.QueryRowContext(ctx, query, args...)

	var incidentRow IncidentRow
	err := row.Scan(
		&incidentRow.ID,
		&incidentRow.ServiceID,
		&incidentRow.StartTime,
		&incidentRow.EndTime,
		&incidentRow.Error,
		&incidentRow.DurationNS,
		&incidentRow.Resolved,
		&incidentRow.CreatedAt,
		&incidentRow.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan incident: %w", err)
	}

	return o.rowToIncident(&incidentRow), nil
}

type FindIncidentsParams struct {
	// Search by service id or incident id
	Search    string
	ID        string
	ServiceID string
	Resolved  *bool
	StartTime *time.Time
	EndTime   *time.Time
	Page      *uint32
	PageSize  *uint32
}

func findIncidentsBuilder(params FindIncidentsParams, col ...string) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(col...)
	sb.From("incidents i")

	if params.ID != "" {
		sb.Where(sb.Equal("i.id", params.ID))
	}

	if params.ServiceID != "" {
		sb.Where(sb.Equal("i.service_id", params.ServiceID))
	}

	if params.Search != "" {
		likeCondition := fmt.Sprintf("%%%s%%", params.Search)
		sb.Where(sb.Or(
			sb.Like("i.id", likeCondition),
			sb.Like("i.service_id", likeCondition),
		))
	}

	if params.Resolved != nil {
		sb.Where(sb.Equal("i.resolved", *params.Resolved))
	}

	if params.StartTime != nil {
		sb.Where(sb.GreaterEqualThan("i.start_time", *params.StartTime))
	}

	if params.EndTime != nil {
		sb.Where(sb.LessEqualThan("i.end_time", *params.EndTime))
	}

	return sb
}

// FindIncidents finds incidents
func (o *ORMStorage) FindIncidents(ctx context.Context, params FindIncidentsParams) (dbutils.FindResponseWithCount[*Incident], error) {
	sb := findIncidentsBuilder(params,
		"i.id",
		"i.service_id",
		"i.start_time",
		"i.end_time",
		"i.error",
		"i.duration_ns",
		"i.resolved",
		"i.created_at",
		"i.updated_at",
	)
	sb.OrderBy("i.start_time").Desc()

	res := dbutils.FindResponseWithCount[*Incident]{}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize)
	if err != nil {
		return res, fmt.Errorf("failed to apply pagination: %w", err)
	}

	sb.Limit(int(limit)).Offset(int(offset))

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return res, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	incidents := []*Incident{}
	for rows.Next() {
		var incidentRow IncidentRow
		err := rows.Scan(
			&incidentRow.ID,
			&incidentRow.ServiceID,
			&incidentRow.StartTime,
			&incidentRow.EndTime,
			&incidentRow.Error,
			&incidentRow.DurationNS,
			&incidentRow.Resolved,
			&incidentRow.CreatedAt,
			&incidentRow.UpdatedAt,
		)
		if err != nil {
			return res, fmt.Errorf("failed to scan incident: %w", err)
		}

		incidents = append(incidents, o.rowToIncident(&incidentRow))
	}

	if err := rows.Err(); err != nil {
		return res, fmt.Errorf("error iterating rows: %w", err)
	}

	// Get total count of incidents
	var totalCount uint32
	countBuilder := findIncidentsBuilder(params, "COUNT(*)")

	countQuery, countArgs := countBuilder.Build()
	err = o.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return res, fmt.Errorf("failed to count incidents: %w", err)
	}

	res.Count = totalCount
	res.Items = incidents

	return res, nil
}

// FindIncidents finds incidents
func (o *ORMStorage) IncidentsCount(ctx context.Context, params FindIncidentsParams) (uint32, error) {
	// Get total count of incidents
	var totalCount uint32
	countBuilder := findIncidentsBuilder(params, "COUNT(*)")

	countQuery, countArgs := countBuilder.Build()
	err := o.db.QueryRowContext(ctx, countQuery, countArgs...).Scan(&totalCount)
	if err != nil {
		return 0, fmt.Errorf("failed to count incidents: %w", err)
	}

	return totalCount, nil
}

// ResolveAllIncidents resolves all incidents for a service
func (o *ORMStorage) ResolveAllIncidents(ctx context.Context, serviceID string) ([]*Incident, error) {
	if serviceID == "" {
		return nil, fmt.Errorf("serviceID is required")
	}

	items, err := o.FindIncidents(ctx, FindIncidentsParams{
		ServiceID: serviceID,
		Resolved:  utils.Pointer(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find incidents: %w", err)
	}

	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	for _, item := range items.Items {
		now := time.Now()

		// Update all incidents for the service to resolved
		ub := sqlbuilder.NewUpdateBuilder()

		ub.Update("incidents").
			Set(
				ub.Assign("resolved", true),
				ub.Assign("end_time", now),
				ub.Assign("duration_ns", now.Sub(item.StartTime)),
				ub.Assign("updated_at", now),
			).
			Where(
				ub.Equal("id", item.ID),
			)

		sql, args := ub.Build()
		if _, err := tx.ExecContext(ctx, sql, args...); err != nil {
			return nil, fmt.Errorf("failed to resolve incidents: %w", err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	resolvedIncidents := []*Incident{}
	for _, item := range items.Items {
		incident, err := o.GetIncidentByID(ctx, item.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get incident by ID: %w", err)
		}

		resolvedIncidents = append(resolvedIncidents, incident)
	}

	return resolvedIncidents, nil
}

// CreateIncident creates a new incident using ORM with retry logic
func (o *ORMStorage) CreateIncident(ctx context.Context, incident *Incident) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("incidents")
	ib.Cols("id", "service_id", "start_time", "end_time", "error", "duration_ns", "resolved")

	ib.Values(
		incident.ID,
		incident.ServiceID,
		incident.StartTime,
		incident.EndTime,
		incident.Error,
		durationToNS(incident.Duration),
		incident.Resolved,
	)

	sql, args := ib.Build()
	_, err := o.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to create incident: %w", err)
	}

	return nil
}

// UpdateIncident updates an existing incident using ORM with retry logic
func (o *ORMStorage) UpdateIncident(ctx context.Context, incident *Incident) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("incidents")
	ub.Set(
		ub.Assign("service_id", incident.ServiceID),
		ub.Assign("start_time", incident.StartTime),
		ub.Assign("end_time", incident.EndTime),
		ub.Assign("error", incident.Error),
		ub.Assign("duration_ns", durationToNS(incident.Duration)),
		ub.Assign("resolved", incident.Resolved),
		ub.Assign("updated_at", time.Now()),
	)
	ub.Where(ub.Equal("id", incident.ID))

	sql, args := ub.Build()
	_, err := o.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	return nil
}

// DeleteIncident deletes an incident by ID using ORM with retry logic
func (o *ORMStorage) DeleteIncident(ctx context.Context, incidentID string) error {
	db := sqlbuilder.NewDeleteBuilder()
	db.DeleteFrom("incidents")
	db.Where(db.Equal("id", incidentID))

	sql, args := db.Build()
	_, err := o.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to delete incident: %w", err)
	}

	return nil
}

type GetIncidentsStatsByDateRangeItem struct {
	Date          time.Time     `json:"date"`
	Count         int64         `json:"count"`
	AvgDuration   time.Duration `json:"avg_duration"`
	TotalDuration time.Duration `json:"total_duration"`
}

type GetIncidentsStatsByDateRangeData []GetIncidentsStatsByDateRangeItem

// GetIncidentsStatsByDateRange retrieves the stats of incidents within a specific date range
func (o *ORMStorage) GetIncidentsStatsByDateRange(ctx context.Context, startTime, endTime time.Time) (GetIncidentsStatsByDateRangeData, error) {
	// Generate date series for the range
	var result GetIncidentsStatsByDateRangeData

	// Iterate through each day in the range
	for d := startTime; !d.After(endTime); d = d.AddDate(0, 0, 1) {
		// Get start and end of the day
		dayStart := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		dayEnd := time.Date(d.Year(), d.Month(), d.Day(), 23, 59, 59, 999999999, d.Location())

		// Query incidents for this specific date range
		sb := sqlbuilder.NewSelectBuilder()
		sb.Select(
			"COUNT(*) as count",
			"AVG(CASE WHEN resolved = true THEN duration_ns ELSE (strftime('%s', 'now') - strftime('%s', start_time)) * 1000000000 END) as avg_duration",
			"SUM(CASE WHEN resolved = true THEN duration_ns ELSE (strftime('%s', 'now') - strftime('%s', start_time)) * 1000000000 END) as total_duration",
		)
		sb.From("incidents")
		sb.Where(sb.GreaterEqualThan("start_time", dayStart))
		sb.Where(sb.LessEqualThan("start_time", dayEnd))

		sql, args := sb.Build()
		row := o.db.QueryRowContext(ctx, sql, args...)

		var count int64
		var avgDurationNS *float64
		var totalDurationNS *float64

		if err := row.Scan(&count, &avgDurationNS, &totalDurationNS); err != nil {
			return nil, fmt.Errorf("failed to scan incident stats for date %s: %w", d.Format("2006-01-02"), err)
		}

		item := GetIncidentsStatsByDateRangeItem{
			Date:  d,
			Count: count,
		}

		// Convert nanoseconds to duration
		if avgDurationNS != nil {
			item.AvgDuration = time.Duration(int64(*avgDurationNS))
		}

		if totalDurationNS != nil {
			item.TotalDuration = time.Duration(int64(*totalDurationNS))
		}

		result = append(result, item)
	}

	return result, nil
}
