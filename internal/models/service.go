package models

import (
	"fmt"
	"time"
)

type ServiceProtocolType string

const (
	ServiceProtocolTypeUnknown ServiceProtocolType = "unknown"
	ServiceProtocolTypeHTTP    ServiceProtocolType = "http"
	ServiceProtocolTypeTCP     ServiceProtocolType = "tcp"
	ServiceProtocolTypeGRPC    ServiceProtocolType = "grpc"
)

// Validate checks if the ServiceProtocolType is valid
func (s ServiceProtocolType) Validate() error {
	switch s {
	case ServiceProtocolTypeHTTP, ServiceProtocolTypeTCP, ServiceProtocolTypeGRPC:
		return nil
	default:
		return fmt.Errorf("invalid service protocol type: %s", s)
	}
}

// String returns the string representation of the ServiceProtocolType
func (s ServiceProtocolType) String() string {
	return string(s)
}

// ServiceStatus represents the current status of a service
type ServiceStatus string

const (
	ServiceStatusUnknown ServiceStatus = "unknown"
	ServiceStatusUp      ServiceStatus = "up"
	ServiceStatusDown    ServiceStatus = "down"
)

// Validate checks if the ServiceStatus is valid
func (s ServiceStatus) Validate() error {
	switch s {
	case ServiceStatusUnknown, ServiceStatusUp, ServiceStatusDown:
		return nil
	default:
		return fmt.Errorf("invalid service status: %s", s)
	}
}

func (s ServiceStatus) String() string {
	return string(s)
}

type ServiceFullView struct {
	ID                     string              `json:"id"`
	Name                   string              `json:"name"`
	Protocol               ServiceProtocolType `json:"protocol"`
	Interval               int64               `json:"interval"`
	Timeout                int64               `json:"timeout"`
	Retries                int64               `json:"retries"`
	Tags                   []string            `json:"tags"`
	Config                 map[string]any      `json:"config"`
	IsNotificationsEnabled bool                `json:"is_notifications_enabled"`
	IsEnabled              bool                `json:"is_enabled"`
	CreatedAt              time.Time           `json:"created_at"`
	UpdatedAt              time.Time           `json:"updated_at"`
	ActiveIncidents        int                 `json:"active_incidents,omitempty"`
	TotalIncidents         int                 `json:"total_incidents,omitempty"`
	Status                 ServiceStatus       `json:"status"`
	LastCheck              *time.Time          `json:"last_check,omitempty"`
	LastError              *string             `json:"last_error,omitempty"`
	ConsecutiveFails       int                 `json:"consecutive_fails"`
	ConsecutiveSuccess     int                 `json:"consecutive_success"`
	TotalChecks            int                 `json:"total_checks"`
	AvgResponseTime        *int64              `json:"avg_response_time"`
}

// GetConfig returns config value by key or default if not set
// func (s *Service) GetConfig() (map[string]any, error) {
// 	var config map[string]any
// 	if err := json.Unmarshal([]byte(s.Config), &config); err != nil {
// 		return nil, err
// 	}
// 	return config, nil
// }
