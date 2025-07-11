package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/huandu/go-sqlbuilder"
)

// ORMStorage provides ORM-like functionality using go-sqlbuilder
type ORMStorage struct {
	db *sql.DB
}

// NewORMStorage creates a new ORM storage instance
func NewORMStorage(db *sql.DB) *ORMStorage {
	return &ORMStorage{db: db}
}

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

// ServiceRow represents a database row for services
type ServiceRow struct {
	ID        string    `db:"id"`
	Name      string    `db:"name"`
	Protocol  string    `db:"protocol"`
	Interval  string    `db:"interval"`
	Timeout   string    `db:"timeout"`
	Retries   int       `db:"retries"`
	Tags      string    `db:"tags"`
	Config    string    `db:"config"`
	IsEnabled bool      `db:"is_enabled"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// QueryIncidents creates a query builder for incidents
func (o *ORMStorage) QueryIncidents() *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "service_id", "start_time", "end_time", "error", "duration_ns", "resolved")
	sb.From("incidents")
	return sb
}

// FindIncidentByID finds an incident by ID using ORM
func (o *ORMStorage) FindIncidentByID(ctx context.Context, serviceID, incidentID string) (*Incident, error) {
	sb := o.QueryIncidents()
	sb.Where(sb.Equal("id", incidentID), sb.Equal("service_id", serviceID))

	sql, args := sb.Build()
	row := o.db.QueryRowContext(ctx, sql, args...)

	var incidentRow IncidentRow
	err := row.Scan(
		&incidentRow.ID,
		&incidentRow.ServiceID,
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

// FindIncidentsByService finds incidents by service ID using ORM
func (o *ORMStorage) FindIncidentsByService(ctx context.Context, serviceID string) ([]*Incident, error) {
	sb := o.QueryIncidents()
	sb.Where(sb.Equal("service_id", serviceID))
	sb.OrderBy("start_time").Desc()

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

	return incidents, nil
}

// FindActiveIncidents finds all active incidents using ORM
func (o *ORMStorage) FindActiveIncidents(ctx context.Context) ([]*Incident, error) {
	sb := o.QueryIncidents()
	sb.Where(sb.Equal("resolved", false))
	sb.OrderBy("start_time").Desc()

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active incidents: %w", err)
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

	return incidents, nil
}

// FindRecentIncidents finds recent incidents with limit using ORM
func (o *ORMStorage) FindRecentIncidents(ctx context.Context, limit int) ([]*Incident, error) {
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

	return incidents, nil
}

// CreateIncident creates a new incident using ORM with retry logic
func (o *ORMStorage) CreateIncident(ctx context.Context, incident *Incident) error {
	return o.retryOnBusy(ctx, func() error {
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
	})
}

// UpdateIncident updates an existing incident using ORM with retry logic
func (o *ORMStorage) UpdateIncident(ctx context.Context, incident *Incident) error {
	return o.retryOnBusy(ctx, func() error {
		ub := sqlbuilder.NewUpdateBuilder()
		ub.Update("incidents")
		ub.Set(
			ub.Assign("service_id", incident.ServiceID),
			ub.Assign("start_time", incident.StartTime),
			ub.Assign("end_time", incident.EndTime),
			ub.Assign("error", incident.Error),
			ub.Assign("duration_ns", durationToNS(incident.Duration)),
			ub.Assign("resolved", incident.Resolved),
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
	})
}

// DeleteIncident deletes an incident by ID using ORM with retry logic
func (o *ORMStorage) DeleteIncident(ctx context.Context, incidentID string) error {
	return o.retryOnBusy(ctx, func() error {
		db := sqlbuilder.NewDeleteBuilder()
		db.DeleteFrom("incidents")
		db.Where(db.Equal("id", incidentID))

		sql, args := db.Build()
		result, err := o.db.ExecContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to delete incident: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("incident not found")
		}

		return nil
	})
}

// GetServiceStatsWithORM calculates statistics for a service using ORM
func (o *ORMStorage) GetServiceStatsWithORM(ctx context.Context, serviceID string, since time.Time) (*ServiceStats, error) {
	// Get all incidents for the service since the specified time
	sb := o.QueryIncidents()
	sb.Where(sb.Equal("service_id", serviceID), sb.GE("start_time", since))

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
	period := time.Since(since)
	uptimePercentage := 100.0
	if period > 0 {
		uptimePercentage = 100.0 - (float64(totalDowntime) / float64(period) * 100.0)
		if uptimePercentage < 0 {
			uptimePercentage = 0
		}
	}

	// Get average response time from service state
	avgResponseTime := time.Duration(0)
	serviceState, err := o.GetServiceState(ctx, serviceID)
	if err != nil {
		// If service state not found, return stats without response time
		return &ServiceStats{
			ServiceID:        serviceID,
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
		ServiceID:        serviceID,
		TotalIncidents:   totalIncidents,
		TotalDowntime:    totalDowntime,
		UptimePercentage: uptimePercentage,
		Period:           period,
		AvgResponseTime:  avgResponseTime,
	}, nil
}

// GetAllServicesIncidentStats gets incident statistics for all services using ORM
func (o *ORMStorage) GetAllServicesIncidentStats(ctx context.Context) ([]*ServiceIncidentStats, error) {
	// Query to get incident statistics for all services
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(
		"service_id",
		"COUNT(*) as total_incidents",
		"SUM(CASE WHEN resolved = 0 THEN 1 ELSE 0 END) as active_incidents",
	)
	sb.From("incidents")
	sb.GroupBy("service_id")

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query incident statistics: %w", err)
	}
	defer rows.Close()

	stats := []*ServiceIncidentStats{}
	for rows.Next() {
		var serviceID string
		var totalIncidents, activeIncidents int
		err := rows.Scan(&serviceID, &totalIncidents, &activeIncidents)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident statistics: %w", err)
		}

		stats = append(stats, &ServiceIncidentStats{
			ServiceID:       serviceID,
			ActiveIncidents: activeIncidents,
			TotalIncidents:  totalIncidents,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return stats, nil
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

// QueryServices creates a query builder for services
func (o *ORMStorage) QueryServices() *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()
	sb.Select("id", "name", "protocol", "interval", "timeout", "retries", "tags", "config", "is_enabled")
	sb.From("services")
	return sb
}

// FindServiceByID finds a service by ID using ORM
func (o *ORMStorage) FindServiceByID(ctx context.Context, id string) (*Service, error) {
	sb := o.QueryServices()
	sb.Where(sb.Equal("id", id))

	sql, args := sb.Build()
	row := o.db.QueryRowContext(ctx, sql, args...)

	var serviceRow ServiceRow
	err := row.Scan(
		&serviceRow.ID,
		&serviceRow.Name,
		&serviceRow.Protocol,
		&serviceRow.Interval,
		&serviceRow.Timeout,
		&serviceRow.Retries,
		&serviceRow.Tags,
		&serviceRow.Config,
		&serviceRow.IsEnabled,
	)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, fmt.Errorf("service not found")
		}
		return nil, fmt.Errorf("failed to scan service: %w", err)
	}

	return o.rowToService(&serviceRow)
}

// FindAllServices finds all services using ORM
func (o *ORMStorage) FindAllServices(ctx context.Context) ([]*Service, error) {
	sb := o.QueryServices()
	sb.OrderBy("name")

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	services := []*Service{}
	for rows.Next() {
		var serviceRow ServiceRow
		err := rows.Scan(
			&serviceRow.ID,
			&serviceRow.Name,
			&serviceRow.Protocol,
			&serviceRow.Interval,
			&serviceRow.Timeout,
			&serviceRow.Retries,
			&serviceRow.Tags,
			&serviceRow.Config,
			&serviceRow.IsEnabled,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		service, err := o.rowToService(&serviceRow)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return services, nil
}

// FindEnabledServices finds all enabled services using ORM
func (o *ORMStorage) FindEnabledServices(ctx context.Context) ([]*Service, error) {
	sb := o.QueryServices()
	sb.Where(sb.Equal("is_enabled", true))
	sb.OrderBy("name")

	sql, args := sb.Build()
	rows, err := o.db.QueryContext(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query active services: %w", err)
	}
	defer rows.Close()

	services := []*Service{}
	for rows.Next() {
		var serviceRow ServiceRow
		err := rows.Scan(
			&serviceRow.ID,
			&serviceRow.Name,
			&serviceRow.Protocol,
			&serviceRow.Interval,
			&serviceRow.Timeout,
			&serviceRow.Retries,
			&serviceRow.Tags,
			&serviceRow.Config,
			&serviceRow.IsEnabled,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}

		service, err := o.rowToService(&serviceRow)
		if err != nil {
			return nil, err
		}
		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return services, nil
}

// CreateService creates a new service using ORM with retry logic
func (o *ORMStorage) CreateService(ctx context.Context, service *Service) error {
	return o.retryOnBusy(ctx, func() error {
		ib := sqlbuilder.NewInsertBuilder()
		ib.InsertInto("services")
		ib.Cols("id", "name", "protocol", "interval", "timeout", "retries", "tags", "config", "is_enabled")

		tagsJSON, err := json.Marshal(service.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshal tags: %w", err)
		}

		configJSON, err := json.Marshal(service.Config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
		}

		ib.Values(
			service.ID,
			service.Name,
			service.Protocol,
			service.Interval.String(),
			service.Timeout.String(),
			service.Retries,
			string(tagsJSON),
			string(configJSON),
			service.IsEnabled,
		)

		sql, args := ib.Build()

		_, err = o.db.ExecContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}

		return nil
	})
}

// UpdateService updates an existing service using ORM with retry logic
func (o *ORMStorage) UpdateService(ctx context.Context, service *Service) error {
	return o.retryOnBusy(ctx, func() error {
		ub := sqlbuilder.NewUpdateBuilder()
		ub.Update("services")

		tagsJSON, err := json.Marshal(service.Tags)
		if err != nil {
			return fmt.Errorf("failed to marshal tags: %w", err)
		}

		configJSON, err := json.Marshal(service.Config)
		if err != nil {
			return fmt.Errorf("failed to marshal config: %w", err)
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
		}

		// Set all assignments at once
		ub.Set(assignments...)
		ub.Where(ub.Equal("id", service.ID))

		sql, args := ub.Build()

		result, err := o.db.ExecContext(ctx, sql, args...)
		if err != nil {
			return fmt.Errorf("failed to update service: %w", err)
		}

		rowsAffected, err := result.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to get rows affected: %w", err)
		}

		if rowsAffected == 0 {
			return fmt.Errorf("service not found")
		}

		return nil
	})
}

// DeleteIncidentsByService deletes all incidents for a service
func (o *ORMStorage) DeleteIncidentsByService(ctx context.Context, serviceID string) error {
	query := `DELETE FROM incidents WHERE service_id = ?`
	_, err := o.db.ExecContext(ctx, query, serviceID)
	if err != nil {
		return fmt.Errorf("failed to delete incidents: %w", err)
	}
	return nil
}

// DeleteService deletes a service by ID
func (o *ORMStorage) DeleteService(ctx context.Context, id string) error {
	// Start transaction
	tx, err := o.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Will be ignored if tx.Commit() is called

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
		&state.ID, &state.ServiceID, &state.Status, &state.LastCheck, &state.NextCheck,
		&state.LastError, &state.ConsecutiveFails, &state.ConsecutiveSuccess,
		&state.TotalChecks, &state.ResponseTimeNS, &state.CreatedAt, &state.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // No state found
		}
		return nil, fmt.Errorf("failed to get service state: %w", err)
	}

	return &state, nil
}

// UpdateServiceState updates or creates service state
func (o *ORMStorage) UpdateServiceState(ctx context.Context, state *ServiceStateRecord) error {
	if state.ID == "" {
		state.ID = GenerateULID()
	}

	query := `
		INSERT INTO service_states (
			id, service_id, status, last_check, next_check, last_error,
			consecutive_fails, consecutive_success, total_checks, response_time_ns,
			created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(service_id) DO UPDATE SET
			status = excluded.status,
			last_check = excluded.last_check,
			next_check = excluded.next_check,
			last_error = excluded.last_error,
			consecutive_fails = excluded.consecutive_fails,
			consecutive_success = excluded.consecutive_success,
			total_checks = excluded.total_checks,
			response_time_ns = excluded.response_time_ns,
			updated_at = excluded.updated_at
	`

	now := time.Now()
	_, err := o.db.ExecContext(ctx, query,
		state.ID, state.ServiceID, state.Status, state.LastCheck, state.NextCheck,
		state.LastError, state.ConsecutiveFails, state.ConsecutiveSuccess,
		state.TotalChecks, state.ResponseTimeNS, now, now,
	)
	if err != nil {
		return fmt.Errorf("failed to update service state: %w", err)
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

// rowToService converts a ServiceRow to Service
func (o *ORMStorage) rowToService(row *ServiceRow) (*Service, error) {
	// Parse duration strings
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

	return &Service{
		ID:        row.ID,
		Name:      row.Name,
		Protocol:  ServiceProtocolType(row.Protocol),
		Interval:  interval,
		Timeout:   timeout,
		Retries:   row.Retries,
		Tags:      tags,
		Config:    config,
		IsEnabled: row.IsEnabled,
	}, nil
}

// durationToNS converts a duration pointer to nanoseconds
func durationToNS(d *time.Duration) *int64 {
	if d == nil {
		return nil
	}
	ns := d.Nanoseconds()
	return &ns
}

// retryOnBusy retries an operation when SQLite is busy
func (o *ORMStorage) retryOnBusy(ctx context.Context, operation func() error) error {
	maxRetries := 5
	baseDelay := 10 * time.Millisecond

	for attempt := range maxRetries {
		err := operation()
		if err == nil {
			return nil
		}

		// Check if it's a busy error
		if isBusyError(err) {
			if attempt < maxRetries-1 {
				delay := time.Duration(attempt+1) * baseDelay
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(delay):
					continue
				}
			}
		}

		return err
	}

	return fmt.Errorf("max retries exceeded")
}

// isBusyError checks if the error is a SQLite busy error
func isBusyError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()
	return errStr == "database is locked (5) (SQLITE_BUSY)" ||
		errStr == "database is locked (SQLITE_BUSY)" ||
		errStr == "database is locked" ||
		errStr == "database table is locked"
}
