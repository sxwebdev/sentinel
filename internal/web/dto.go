package web

import (
	"time"

	"github.com/sxwebdev/sentinel/internal/monitors"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// DashboardStats represents dashboard statistics
//
//	@Description	Dashboard statistics
type DashboardStats struct {
	TotalServices    int                                 `json:"total_services" example:"10"`
	ServicesUp       int                                 `json:"services_up" example:"8"`
	ServicesDown     int                                 `json:"services_down" example:"1"`
	ServicesUnknown  int                                 `json:"services_unknown" example:"1"`
	Protocols        map[storage.ServiceProtocolType]int `json:"protocols"`
	ActiveIncidents  int                                 `json:"active_incidents" example:"2"`
	AvgResponseTime  int64                               `json:"avg_response_time" example:"150"`
	TotalChecks      int                                 `json:"total_checks" example:"1000"`
	UptimePercentage float64                             `json:"uptime_percentage" example:"95.5"`
	LastCheckTime    *time.Time                          `json:"last_check_time"`
	ChecksPerMinute  int                                 `json:"checks_per_minute" example:"60"`
}

// Incident represents an incident
//
//	@Description	Service incident
type Incident struct {
	ID          string     `json:"id" example:"01HXYZ1234567890ABCDEF"`
	ServiceID   string     `json:"service_id" example:"service-1"`
	ServiceName string     `json:"service_name" example:"Web Server"`
	Status      string     `json:"status" example:"down"`
	Message     string     `json:"message" example:"Connection timeout"`
	StartedAt   time.Time  `json:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	Resolved    bool       `json:"resolved" example:"false"`
	Duration    string     `json:"duration" example:"2h30m"`
}

// ServiceStats represents service statistics
//
//	@Description	Service statistics
type ServiceStats struct {
	ServiceID        string        `json:"service_id" example:"service-1"`
	TotalIncidents   int           `json:"total_incidents" example:"5"`
	TotalDowntime    time.Duration `json:"total_downtime" swaggertype:"primitive,integer" example:"1800000000000"`
	UptimePercentage float64       `json:"uptime_percentage" example:"95.0"`
	Period           time.Duration `json:"period" swaggertype:"primitive,integer" example:"2592000000000000"`
	AvgResponseTime  time.Duration `json:"avg_response_time" swaggertype:"primitive,integer" example:"150000000"`
}

// CreateUpdateServiceRequest represents a request to create or update a service
type CreateUpdateServiceRequest struct {
	Name      string                      `json:"name" example:"Web Server"`
	Protocol  storage.ServiceProtocolType `json:"protocol" example:"http"`
	Interval  uint32                      `json:"interval" swaggertype:"primitive,integer" example:"60000"`
	Timeout   uint32                      `json:"timeout" swaggertype:"primitive,integer" example:"10000"`
	Retries   int                         `json:"retries" example:"5"`
	Tags      []string                    `json:"tags" example:"web,production"`
	Config    monitors.Config             `json:"config"`
	IsEnabled bool                        `json:"is_enabled" example:"true"`
}

// ServiceDTO represents a service for API responses
type ServiceDTO struct {
	ID                 string                      `json:"id" example:"service-1"`
	Name               string                      `json:"name" example:"Web Server"`
	Protocol           storage.ServiceProtocolType `json:"protocol" example:"http"`
	Interval           uint32                      `json:"interval" swaggertype:"primitive,integer" example:"60000"`
	Timeout            uint32                      `json:"timeout" swaggertype:"primitive,integer" example:"10000"`
	Retries            int                         `json:"retries" example:"5"`
	Tags               []string                    `json:"tags" example:"web,production"`
	Config             monitors.Config             `json:"config"`
	IsEnabled          bool                        `json:"is_enabled" example:"true"`
	ActiveIncidents    int                         `json:"active_incidents" example:"2"`
	TotalIncidents     int                         `json:"total_incidents" example:"10"`
	Status             storage.ServiceStatus       `json:"status" example:"up / down / unknown"`
	LastCheck          *time.Time                  `json:"last_check,omitempty" example:"2023-10-01T12:00:00Z"`
	NextCheck          *time.Time                  `json:"next_check,omitempty" example:"2023-10-01T12:05:00Z"`
	LastError          *string                     `json:"last_error,omitempty" example:"Connection timeout"`
	ConsecutiveFails   int                         `json:"consecutive_fails" example:"1"`
	ConsecutiveSuccess int                         `json:"consecutive_success" example:"5"`
	TotalChecks        int                         `json:"total_checks" example:"100"`
	ResponseTime       uint32                      `json:"response_time" swaggertype:"primitive,integer" example:"150000000"`
}

type ServerInfoResponse struct {
	Version         string           `json:"version" example:"1.0.0"`
	CommitHash      string           `json:"commit_hash" example:"abc123def456"`
	BuildDate       string           `json:"build_date" example:"2023-10-01T12:00:00Z"`
	GoVersion       string           `json:"go_version" example:"go1.24.4"`
	SqliteVersion   string           `json:"sqlite_version" example:"3.50.1"`
	OS              string           `json:"os" example:"linux"`
	Arch            string           `json:"arch" example:"amd64"`
	AvailableUpdate *AvailableUpdate `json:"available_update,omitempty"`
}
