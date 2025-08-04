package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/huandu/go-sqlbuilder"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

// ORMStorage provides ORM-like functionality using go-sqlbuilder
type ORMStorage struct {
	db *sql.DB
}

// NewORMStorage creates a new ORM storage instance
func NewORMStorage(db *sql.DB) *ORMStorage {
	return &ORMStorage{db: db}
}

// GetServiceStatsWithORM calculates statistics for a service using ORM
func (o *ORMStorage) GetServiceStatsWithORM(ctx context.Context, params FindIncidentsParams) (*ServiceStats, error) {
	if params.ServiceID == "" || params.StartTime == nil {
		return nil, fmt.Errorf("service ID and start time are required for stats")
	}

	// Get all incidents for the service since the specified time
	sb := findIncidentsBuilder(params,
		"i.id",
		"i.service_id",
		"i.start_time",
		"i.end_time",
		"i.error",
		"i.duration_ns",
		"i.resolved",
	)

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
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
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		incidents = append(incidents, o.rowToIncident(&incidentRow))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	// Calculate statistics
	totalIncidents := len(incidents)
	totalDowntime := time.Duration(0)
	resolvedIncidents := 0

	for _, incident := range incidents {
		if incident.Resolved && incident.Duration != nil {
			totalDowntime += *incident.Duration
			resolvedIncidents++
		}
	}

	// Calculate uptime percentage
	period := time.Since(*params.StartTime)
	uptimePercentage := 100.0
	if period > 0 {
		uptimePercentage = 100.0 - (float64(totalDowntime) / float64(period) * 100.0)
		if uptimePercentage < 0 {
			uptimePercentage = 0
		}
	}

	// Get average response time from service state
	avgResponseTime := time.Duration(0)
	serviceState, err := o.GetServiceState(ctx, params.ServiceID)
	if err != nil {
		// If service state not found, return stats without response time
		return &ServiceStats{
			ServiceID:        params.ServiceID,
			TotalIncidents:   totalIncidents,
			TotalDowntime:    totalDowntime,
			UptimePercentage: uptimePercentage,
			Period:           period,
			AvgResponseTime:  0,
		}, nil
	}
	if serviceState != nil && serviceState.ResponseTimeNS != nil {
		avgResponseTime = time.Duration(*serviceState.ResponseTimeNS)
	}

	return &ServiceStats{
		ServiceID:        params.ServiceID,
		TotalIncidents:   totalIncidents,
		TotalDowntime:    totalDowntime,
		UptimePercentage: uptimePercentage,
		Period:           period,
		AvgResponseTime:  avgResponseTime,
	}, nil
}

// rowToIncident converts an IncidentRow to Incident
func (o *ORMStorage) rowToIncident(row *IncidentRow) *Incident {
	incident := &Incident{
		ID:        row.ID,
		ServiceID: row.ServiceID,
		StartTime: row.StartTime,
		EndTime:   row.EndTime,
		Error:     row.Error,
		Resolved:  row.Resolved,
	}

	if row.DurationNS != nil {
		duration := time.Duration(*row.DurationNS)
		incident.Duration = &duration
	}

	return incident
}

// GetServiceByID finds a service by ID using ORM
func (o *ORMStorage) GetServiceByID(ctx context.Context, id string) (*Service, error) {
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
		"ss.response_time_ns",
	)
	sb.From("services s")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "incidents", "s.id = incidents.service_id")
	sb.JoinWithOption(sqlbuilder.LeftJoin, "service_states ss", "s.id = ss.service_id")
	sb.Where(sb.Equal("s.id", id))
	sb.GroupBy("s.id")

	query, args := sb.Build()
	row := o.db.QueryRowContext(ctx, query, args...)

	var item serviceRow
	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Protocol,
		&item.Interval,
		&item.Timeout,
		&item.Retries,
		&item.Tags,
		&item.Config,
		&item.IsEnabled,
		&item.CreatedAt,
		&item.UpdatedAt,
		&item.TotalIncidents,
		&item.ActiveIncidents,
		&item.Status,
		&item.LastCheck,
		&item.NextCheck,
		&item.LastError,
		&item.ConsecutiveFails,
		&item.ConsecutiveSuccess,
		&item.TotalChecks,
		&item.ResponseTimeNS,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to scan service: %w", err)
	}

	svc, err := rowToService(&item)
	if err != nil {
		return nil, err
	}

	return svc, nil
}

func findServicesBuilder(params FindServicesParams, col ...string) *sqlbuilder.SelectBuilder {
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
			sb.Where(sb.Equal("ss.status", StatusUp))
		case "down":
			sb.Where(sb.Equal("ss.status", StatusDown))
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
			sb.Where(fmt.Sprintf("(%s)", strings.Join(tagConditions, " OR ")))
		}
	}

	return sb
}

type FindServicesParams struct {
	Name      string
	IsEnabled *bool
	Protocol  string
	Tags      []string
	Status    string // e.g. "up", "down"
	OrderBy   string
	Page      *uint32
	PageSize  *uint32
}

// GetAllServices finds all services using ORM
func (o *ORMStorage) FindServices(ctx context.Context, params FindServicesParams) (dbutils.FindResponseWithCount[*Service], error) {
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
		"sum(case when incidents.resolved = 0 then 1 else 0 end) as active_incidents",
		"ss.status",
		"ss.last_check",
		"ss.next_check",
		"ss.last_error",
		"ss.consecutive_fails",
		"ss.consecutive_success",
		"ss.total_checks",
		"ss.response_time_ns",
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

	res := dbutils.FindResponseWithCount[*Service]{}

	limit, offset, err := dbutils.Pagination(params.Page, params.PageSize)
	if err != nil {
		return res, err
	}
	sb.Limit(int(limit)).Offset(int(offset))

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return res, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	services := []*Service{}
	for rows.Next() {
		var item serviceRow
		err := rows.Scan(
			&item.ID,
			&item.Name,
			&item.Protocol,
			&item.Interval,
			&item.Timeout,
			&item.Retries,
			&item.Tags,
			&item.Config,
			&item.IsEnabled,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.TotalIncidents,
			&item.ActiveIncidents,
			&item.Status,
			&item.LastCheck,
			&item.NextCheck,
			&item.LastError,
			&item.ConsecutiveFails,
			&item.ConsecutiveSuccess,
			&item.TotalChecks,
			&item.ResponseTimeNS,
		)
		if err != nil {
			return res, fmt.Errorf("failed to scan service: %w", err)
		}

		svc, err := rowToService(&item)
		if err != nil {
			return res, fmt.Errorf("failed to convert service row: %w", err)
		}

		services = append(services, svc)
	}

	if err := rows.Err(); err != nil {
		return res, fmt.Errorf("error iterating rows: %w", err)
	}

	// Get total count of services
	countQuery := findServicesBuilder(params, "count(*)")
	countQuery.JoinWithOption(sqlbuilder.LeftJoin, "service_states ss", "s.id = ss.service_id")

	countSQL, countArgs := countQuery.Build()

	var totalCount int
	if err := o.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&totalCount); err != nil {
		return res, fmt.Errorf("failed to count services: %w", err)
	}

	res.Count = uint32(totalCount)
	res.Items = services

	return res, nil
}

// CreateService creates a new service using ORM with retry logic
func (o *ORMStorage) CreateService(ctx context.Context, service CreateUpdateServiceRequest) (*Service, error) {
	ib := sqlbuilder.NewInsertBuilder()
	ib.InsertInto("services")
	ib.Cols("id", "name", "protocol", "interval", "timeout", "retries", "tags", "config", "is_enabled")

	tagsJSON, err := json.Marshal(service.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	configJSON, err := json.Marshal(service.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	serviceID := GenerateULID()

	ib.Values(
		serviceID,
		service.Name,
		service.Protocol,
		service.Interval.String(),
		service.Timeout.String(),
		service.Retries,
		string(tagsJSON),
		string(configJSON),
		service.IsEnabled,
	)

	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	sql, args := ib.Build()
	_, err = tx.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}

	nextCheck := time.Now().Add(service.Interval)
	serviceState := &ServiceStateRecord{
		ID:        GenerateULID(),
		ServiceID: serviceID,
		Status:    StatusUnknown,
		NextCheck: &nextCheck,
	}

	if err := o.CreateServiceState(ctx, tx, serviceState); err != nil {
		return nil, fmt.Errorf("failed to create service state: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return o.GetServiceByID(ctx, serviceID)
}

// UpdateService updates an existing service using ORM with retry logic
func (o *ORMStorage) UpdateService(ctx context.Context, id string, service CreateUpdateServiceRequest) (*Service, error) {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("services")

	tagsJSON, err := json.Marshal(service.Tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	configJSON, err := json.Marshal(service.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config: %w", err)
	}

	// Prepare all fields for update
	assignments := []string{
		ub.Assign("name", service.Name),
		ub.Assign("protocol", service.Protocol),
		ub.Assign("interval", service.Interval.String()),
		ub.Assign("timeout", service.Timeout.String()),
		ub.Assign("retries", service.Retries),
		ub.Assign("tags", string(tagsJSON)),
		ub.Assign("config", string(configJSON)),
		ub.Assign("is_enabled", service.IsEnabled),
		ub.Assign("updated_at", time.Now()),
	}

	// Set all assignments at once
	ub.Set(assignments...)
	ub.Where(ub.Equal("id", id))

	sql, args := ub.Build()
	result, err := o.db.ExecContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("service not found")
	}

	return o.GetServiceByID(ctx, id)
}

// DeleteService deletes a service by ID
func (o *ORMStorage) DeleteService(ctx context.Context, id string) error {
	// Start transaction
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete related incidents first
	incidentsQuery := `DELETE FROM incidents WHERE service_id = ?`
	_, err = tx.ExecContext(ctx, incidentsQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete incidents: %w", err)
	}

	// Delete service state
	stateQuery := `DELETE FROM service_states WHERE service_id = ?`
	_, err = tx.ExecContext(ctx, stateQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete service state: %w", err)
	}

	// Delete the service
	serviceQuery := `DELETE FROM services WHERE id = ?`
	result, err := tx.ExecContext(ctx, serviceQuery, id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("service not found")
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Service state management methods

// GetServiceState gets service state by service ID
func (o *ORMStorage) GetServiceState(ctx context.Context, serviceID string) (*ServiceStateRecord, error) {
	query := `
		SELECT id, service_id, status, last_check, next_check, last_error, 
		       consecutive_fails, consecutive_success, total_checks, response_time_ns,
		       created_at, updated_at
		FROM service_states 
		WHERE service_id = ?
	`

	var state ServiceStateRecord
	err := o.db.QueryRowContext(ctx, query, serviceID).Scan(
		&state.ID,
		&state.ServiceID,
		&state.Status,
		&state.LastCheck,
		&state.NextCheck,
		&state.LastError,
		&state.ConsecutiveFails,
		&state.ConsecutiveSuccess,
		&state.TotalChecks,
		&state.ResponseTimeNS,
		&state.CreatedAt,
		&state.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get service state: %w", err)
	}

	return &state, nil
}

// CreateServiceState creates a new service state
func (o *ORMStorage) CreateServiceState(ctx context.Context, tx *sql.Tx, state *ServiceStateRecord) error {
	query := `
		INSERT INTO service_states (
			id, service_id, status, last_check, next_check, last_error,
			consecutive_fails, consecutive_success, total_checks, response_time_ns
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := tx.ExecContext(ctx, query,
		state.ID,
		state.ServiceID,
		state.Status,
		state.LastCheck,
		state.NextCheck,
		state.LastError,
		state.ConsecutiveFails,
		state.ConsecutiveSuccess,
		state.TotalChecks,
		state.ResponseTimeNS,
	)
	if err != nil {
		return fmt.Errorf("failed to create service state: %w", err)
	}
	return nil
}

// UpdateServiceState updates or creates service state
func (o *ORMStorage) UpdateServiceState(ctx context.Context, params *ServiceStateRecord) error {
	ub := sqlbuilder.NewUpdateBuilder()
	ub.Update("service_states")
	ub.Set(
		ub.Assign("status", params.Status),
		ub.Assign("last_check", params.LastCheck),
		ub.Assign("next_check", params.NextCheck),
		ub.Assign("last_error", params.LastError),
		ub.Assign("consecutive_fails", params.ConsecutiveFails),
		ub.Assign("consecutive_success", params.ConsecutiveSuccess),
		ub.Assign("total_checks", params.TotalChecks),
		ub.Assign("response_time_ns", params.ResponseTimeNS),
		ub.Assign("updated_at", time.Now()),
	)

	ub.Where(ub.Equal("id", params.ID))

	query, args := ub.Build()
	if _, err := o.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}

	return nil
}

// GetAllServiceStates gets all service states
func (o *ORMStorage) GetAllServiceStates(ctx context.Context) ([]*ServiceStateRecord, error) {
	query := `
		SELECT id, service_id, status, last_check, next_check, last_error,
		       consecutive_fails, consecutive_success, total_checks, response_time_ns,
		       created_at, updated_at
		FROM service_states
		ORDER BY updated_at DESC
	`

	rows, err := o.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query service states: %w", err)
	}
	defer rows.Close()

	states := []*ServiceStateRecord{}
	for rows.Next() {
		var state ServiceStateRecord
		err := rows.Scan(
			&state.ID, &state.ServiceID, &state.Status, &state.LastCheck, &state.NextCheck,
			&state.LastError, &state.ConsecutiveFails, &state.ConsecutiveSuccess,
			&state.TotalChecks, &state.ResponseTimeNS, &state.CreatedAt, &state.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service state: %w", err)
		}
		states = append(states, &state)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating service states: %w", err)
	}

	return states, nil
}

// DeleteServiceState deletes service state by service ID
func (o *ORMStorage) DeleteServiceState(ctx context.Context, serviceID string) error {
	query := `DELETE FROM service_states WHERE service_id = ?`
	_, err := o.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete service state: %w", err)
	}
	return nil
}

func (o *ORMStorage) GetAllTags(ctx context.Context) ([]string, error) {
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

func (o *ORMStorage) GetAllTagsWithCount(ctx context.Context) (map[string]int, error) {
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

// rowToService converts a ServiceRow to Service
func rowToService(row *serviceRow) (*Service, error) {
	interval, err := time.ParseDuration(row.Interval)
	if err != nil {
		return nil, fmt.Errorf("failed to parse interval: %w", err)
	}

	timeout, err := time.ParseDuration(row.Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to parse timeout: %w", err)
	}

	var tags []string
	if err := json.Unmarshal([]byte(row.Tags), &tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal([]byte(row.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	svc := &Service{
		ID:                 row.ID,
		Name:               row.Name,
		Protocol:           ServiceProtocolType(row.Protocol),
		Interval:           interval,
		Timeout:            timeout,
		Retries:            row.Retries,
		Tags:               tags,
		Config:             config,
		IsEnabled:          row.IsEnabled,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
		TotalIncidents:     row.TotalIncidents,
		ActiveIncidents:    row.ActiveIncidents,
		Status:             row.Status,
		LastCheck:          row.LastCheck,
		NextCheck:          row.NextCheck,
		LastError:          row.LastError,
		ConsecutiveFails:   row.ConsecutiveFails,
		ConsecutiveSuccess: row.ConsecutiveSuccess,
		TotalChecks:        row.TotalChecks,
	}

	if row.ResponseTimeNS != nil {
		svc.ResponseTime = utils.Pointer(time.Duration(*row.ResponseTimeNS))
	}

	return svc, nil
}

// durationToNS converts a duration pointer to nanoseconds
func durationToNS(d *time.Duration) *int64 {
	if d == nil {
		return nil
	}
	ns := d.Nanoseconds()
	return &ns
}
