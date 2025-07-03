package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/config"
)

// ORMStorage provides ORM-like functionality using go-sqlbuilder
type ORMStorage struct {
	db *sql.DB
}

// NewORMStorage creates a new ORM storage instance
func NewORMStorage(db *sql.DB) *ORMStorage {
	return &ORMStorage{db: db}
}

// QueryIncidents creates a query builder for incidents
func (o *ORMStorage) QueryIncidents() *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "service_name", "start_time", "end_time", "error", "duration_ns", "resolved")
	sb.From("incidents")
	return sb
}

// FindIncidentByID finds an incident by ID using ORM
func (o *ORMStorage) FindIncidentByID(ctx context.Context, serviceID, incidentID string) (*config.Incident, error) {
	sb := o.QueryIncidents()
	sb.Where(sb.Equal("id", incidentID), sb.Equal("service_name", serviceID))

	sql, args := sb.Build()
	row := o.db.QueryRowContext(ctx, sql, args...)

	var incidentRow IncidentRow
	err := row.Scan(
		&incidentRow.ID,
		&incidentRow.ServiceName,
		&incidentRow.StartTime,
		&incidentRow.EndTime,
		&incidentRow.Error,
		&incidentRow.DurationNS,
		&incidentRow.Resolved,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("incident not found")
		}
		return nil, fmt.Errorf("failed to scan incident: %w", err)
	}

	return o.rowToIncident(&incidentRow), nil
}

// FindIncidentsByService finds incidents by service name using ORM
func (o *ORMStorage) FindIncidentsByService(ctx context.Context, serviceName string) ([]*config.Incident, error) {
	sb := o.QueryIncidents()
	sb.Where(sb.Equal("service_name", serviceName))
	sb.OrderBy("start_time").Desc()

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*config.Incident
	for rows.Next() {
		var incidentRow IncidentRow
		err := rows.Scan(
			&incidentRow.ID,
			&incidentRow.ServiceName,
			&incidentRow.StartTime,
			&incidentRow.EndTime,
			&incidentRow.Error,
			&incidentRow.DurationNS,
			&incidentRow.Resolved,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		incidents = append(incidents, o.rowToIncident(&incidentRow))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return incidents, nil
}

// FindActiveIncidents finds all active incidents using ORM
func (o *ORMStorage) FindActiveIncidents(ctx context.Context) ([]*config.Incident, error) {
	sb := o.QueryIncidents()
	sb.Where(sb.Equal("resolved", false))
	sb.OrderBy("start_time").Desc()

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*config.Incident
	for rows.Next() {
		var incidentRow IncidentRow
		err := rows.Scan(
			&incidentRow.ID,
			&incidentRow.ServiceName,
			&incidentRow.StartTime,
			&incidentRow.EndTime,
			&incidentRow.Error,
			&incidentRow.DurationNS,
			&incidentRow.Resolved,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		incidents = append(incidents, o.rowToIncident(&incidentRow))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return incidents, nil
}

// FindRecentIncidents finds recent incidents with limit using ORM
func (o *ORMStorage) FindRecentIncidents(ctx context.Context, limit int) ([]*config.Incident, error) {
	sb := o.QueryIncidents()
	sb.OrderBy("start_time").Desc()

	if limit > 0 {
		sb.Limit(limit)
	}

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []*config.Incident
	for rows.Next() {
		var incidentRow IncidentRow
		err := rows.Scan(
			&incidentRow.ID,
			&incidentRow.ServiceName,
			&incidentRow.StartTime,
			&incidentRow.EndTime,
			&incidentRow.Error,
			&incidentRow.DurationNS,
			&incidentRow.Resolved,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		incidents = append(incidents, o.rowToIncident(&incidentRow))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return incidents, nil
}

// CreateIncident creates a new incident using ORM
func (o *ORMStorage) CreateIncident(ctx context.Context, incident *config.Incident) error {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("incidents")
	ib.Cols("id", "service_name", "start_time", "end_time", "error", "duration_ns", "resolved")

	ib.Values(
		incident.ID,
		incident.ServiceName,
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

// UpdateIncident updates an incident using ORM
func (o *ORMStorage) UpdateIncident(ctx context.Context, incident *config.Incident) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("incidents")

	ub.Set(
		ub.Assign("service_name", incident.ServiceName),
		ub.Assign("start_time", incident.StartTime),
		ub.Assign("end_time", incident.EndTime),
		ub.Assign("error", incident.Error),
		ub.Assign("duration_ns", durationToNS(incident.Duration)),
		ub.Assign("resolved", incident.Resolved),
		ub.Assign("updated_at", time.Now()),
	)
	ub.Where(ub.Equal("id", incident.ID))

	sql, args := ub.Build()
	result, err := o.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to update incident: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("incident not found")
	}

	return nil
}

// GetServiceStatsWithORM calculates statistics using ORM queries
func (o *ORMStorage) GetServiceStatsWithORM(ctx context.Context, serviceName string, since time.Time) (*ServiceStats, error) {
	// Query for incidents since the specified time
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("start_time", "end_time", "duration_ns", "resolved")
	sb.From("incidents")
	sb.Where(
		sb.Equal("service_name", serviceName),
		sb.GE("start_time", since),
	)

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents for stats: %w", err)
	}
	defer rows.Close()

	stats := &ServiceStats{
		ServiceName:    serviceName,
		TotalIncidents: 0,
		TotalDowntime:  0,
		Period:         time.Since(since),
	}

	for rows.Next() {
		var startTime time.Time
		var endTime *time.Time
		var durationNS *int64
		var resolved bool

		err := rows.Scan(&startTime, &endTime, &durationNS, &resolved)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident for stats: %w", err)
		}

		stats.TotalIncidents++

		if durationNS != nil {
			stats.TotalDowntime += time.Duration(*durationNS)
		} else if !resolved {
			// If incident is still active, calculate downtime from start time
			stats.TotalDowntime += time.Since(startTime)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Calculate uptime percentage
	if stats.Period > 0 {
		uptimeRatio := float64(stats.Period-stats.TotalDowntime) / float64(stats.Period)
		stats.UptimePercentage = uptimeRatio * 100
		if stats.UptimePercentage < 0 {
			stats.UptimePercentage = 0
		}
	} else {
		stats.UptimePercentage = 100
	}

	return stats, nil
}

// rowToIncident converts IncidentRow to config.Incident
func (o *ORMStorage) rowToIncident(row *IncidentRow) *config.Incident {
	incident := &config.Incident{
		ID:          row.ID,
		ServiceName: row.ServiceName,
		StartTime:   row.StartTime,
		EndTime:     row.EndTime,
		Error:       row.Error,
		Resolved:    row.Resolved,
	}

	// Convert duration from nanoseconds
	if row.DurationNS != nil {
		duration := time.Duration(*row.DurationNS)
		incident.Duration = &duration
	}

	return incident
}
