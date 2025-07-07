package web

import (
	"time"
)

// ErrorResponse represents an error response
// @Description Error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error description"`
}

// SuccessResponse represents a successful response
// @Description Successful response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// DashboardStats represents dashboard statistics
// @Description Dashboard statistics
type DashboardStats struct {
	TotalServices    int            `json:"total_services" example:"10"`
	ServicesUp       int            `json:"services_up" example:"8"`
	ServicesDown     int            `json:"services_down" example:"1"`
	ServicesUnknown  int            `json:"services_unknown" example:"1"`
	Protocols        map[string]int `json:"protocols"`
	RecentIncidents  int            `json:"recent_incidents" example:"5"`
	ActiveIncidents  int            `json:"active_incidents" example:"2"`
	AvgResponseTime  int64          `json:"avg_response_time" example:"150"`
	TotalChecks      int            `json:"total_checks" example:"1000"`
	UptimePercentage float64        `json:"uptime_percentage" example:"95.5"`
	LastCheckTime    *time.Time     `json:"last_check_time"`
	ChecksPerMinute  int            `json:"checks_per_minute" example:"60"`
}

// Incident represents an incident
// @Description Service incident
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
// @Description Service statistics
type ServiceStats struct {
	ServiceID        string     `json:"service_id" example:"service-1"`
	TotalChecks      int        `json:"total_checks" example:"1000"`
	SuccessfulChecks int        `json:"successful_checks" example:"950"`
	FailedChecks     int        `json:"failed_checks" example:"50"`
	UptimePercent    float64    `json:"uptime_percent" example:"95.0"`
	AvgResponseTime  int64      `json:"avg_response_time" example:"150"`
	LastCheck        time.Time  `json:"last_check"`
	LastIncident     *time.Time `json:"last_incident"`
}
