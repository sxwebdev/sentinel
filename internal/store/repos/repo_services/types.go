package repo_services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
)

type itemViewRow struct {
	ID                     string               `db:"id"`
	Name                   string               `db:"name"`
	Protocol               string               `db:"protocol"`
	Interval               int64                `db:"interval"`
	Timeout                int64                `db:"timeout"`
	Retries                int64                `db:"retries"`
	Tags                   string               `db:"tags"`
	Config                 string               `db:"config"`
	IsEnabled              bool                 `db:"is_enabled"`
	IsNotificationsEnabled bool                 `db:"is_notifications_enabled"`
	CreatedAt              time.Time            `db:"created_at"`
	UpdatedAt              time.Time            `db:"updated_at"`
	ActiveIncidents        int                  `db:"active_incidents"`
	TotalIncidents         int                  `db:"total_incidents"`
	Status                 models.ServiceStatus `db:"status"`
	LastCheck              *time.Time           `db:"last_check"`
	LastError              *string              `db:"last_error"`
	AvgResponseTime        *int64               `db:"avg_response_time"`
	ConsecutiveFails       int                  `db:"consecutive_fails"`
	ConsecutiveSuccess     int                  `db:"consecutive_success"`
	TotalChecks            int                  `db:"total_checks"`
}

// rowToModel converts a ServiceRow to Service
func rowToModel(row *itemViewRow) (*models.ServiceFullView, error) {
	var tags []string
	if err := json.Unmarshal([]byte(row.Tags), &tags); err != nil {
		return nil, fmt.Errorf("failed to unmarshal tags: %w", err)
	}

	var config map[string]any
	if err := json.Unmarshal([]byte(row.Config), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	svc := &models.ServiceFullView{
		ID:                     row.ID,
		Name:                   row.Name,
		Protocol:               models.ServiceProtocolType(row.Protocol),
		Interval:               row.Interval,
		Timeout:                row.Timeout,
		Retries:                row.Retries,
		Tags:                   tags,
		Config:                 config,
		IsNotificationsEnabled: row.IsNotificationsEnabled,
		IsEnabled:              row.IsEnabled,
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
		TotalIncidents:         row.TotalIncidents,
		ActiveIncidents:        row.ActiveIncidents,
		Status:                 row.Status,
		LastCheck:              row.LastCheck,
		LastError:              row.LastError,
		ConsecutiveFails:       row.ConsecutiveFails,
		ConsecutiveSuccess:     row.ConsecutiveSuccess,
		TotalChecks:            row.TotalChecks,
		AvgResponseTime:        row.AvgResponseTime,
	}

	return svc, nil
}

// durationToNS converts a duration pointer to nanoseconds
// func durationToNS(d *time.Duration) *int64 {
// 	if d == nil {
// 		return nil
// 	}
// 	ns := d.Nanoseconds()
// 	return &ns
// }
