package repo_services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/utils"
)

type itemViewRow struct {
	ID                 string
	Name               string
	Protocol           string
	Interval           string
	Timeout            string
	Retries            int
	Tags               string
	Config             string
	IsEnabled          bool
	CreatedAt          time.Time
	UpdatedAt          time.Time
	ActiveIncidents    int
	TotalIncidents     int
	Status             models.ServiceStatus
	LastCheck          *time.Time
	NextCheck          *time.Time
	LastError          *string
	ConsecutiveFails   int
	ConsecutiveSuccess int
	TotalChecks        int
	ResponseTimeNS     *int64
}

// rowToModel converts a ServiceRow to Service
func rowToModel(row *itemViewRow) (*models.ServiceFullView, error) {
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

	svc := &models.ServiceFullView{
		ID:                 row.ID,
		Name:               row.Name,
		Protocol:           models.ServiceProtocolType(row.Protocol),
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
// func durationToNS(d *time.Duration) *int64 {
// 	if d == nil {
// 		return nil
// 	}
// 	ns := d.Nanoseconds()
// 	return &ns
// }
