package models

import (
	"encoding/json"
	"time"
)

type ServiceProtocolType string

const (
	ServiceProtocolTypeHTTP ServiceProtocolType = "http"
	ServiceProtocolTypeTCP  ServiceProtocolType = "tcp"
	ServiceProtocolTypeGRPC ServiceProtocolType = "grpc"
)

// ServiceStatus represents the current status of a service
type ServiceStatus string

const (
	StatusUnknown     ServiceStatus = "unknown"
	StatusUp          ServiceStatus = "up"
	StatusDown        ServiceStatus = "down"
	StatusMaintenance ServiceStatus = "maintenance"
)

func (s ServiceStatus) String() string {
	return string(s)
}

type ServiceFullView struct {
	ID                 string              `json:"id"`
	Name               string              `json:"name"`
	Protocol           ServiceProtocolType `json:"protocol"`
	Interval           time.Duration       `json:"interval" swaggertype:"primitive,integer"`
	Timeout            time.Duration       `json:"timeout" swaggertype:"primitive,integer"`
	Retries            int64               `json:"retries"`
	Tags               []string            `json:"tags"`
	Config             map[string]any      `json:"config"`
	IsEnabled          bool                `json:"is_enabled"`
	CreatedAt          time.Time           `json:"created_at"`
	UpdatedAt          time.Time           `json:"updated_at"`
	ActiveIncidents    int                 `json:"active_incidents,omitempty"`
	TotalIncidents     int                 `json:"total_incidents,omitempty"`
	Status             ServiceStatus       `json:"status"`
	LastCheck          *time.Time          `json:"last_check,omitempty"`
	NextCheck          *time.Time          `json:"next_check,omitempty"`
	LastError          *string             `json:"last_error,omitempty"`
	ConsecutiveFails   int                 `json:"consecutive_fails"`
	ConsecutiveSuccess int                 `json:"consecutive_success"`
	TotalChecks        int                 `json:"total_checks"`
	ResponseTime       *time.Duration      `json:"response_time" swaggertype:"primitive,integer"`
}

// GetConfig returns config value by key or default if not set
func (s *Service) GetConfig() (map[string]any, error) {
	var config map[string]any
	if err := json.Unmarshal([]byte(s.Config), &config); err != nil {
		return nil, err
	}
	return config, nil
}
