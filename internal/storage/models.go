package storage

import (
	"context"
	"encoding/json"
	"time"

	"github.com/oklog/ulid/v2"
)

// Storage defines the interface for incident storage
type Storage interface {
	// Incident management
	SaveIncident(ctx context.Context, incident *Incident) error
	GetIncident(ctx context.Context, serviceID, incidentID string) (*Incident, error)
	UpdateIncident(ctx context.Context, incident *Incident) error
	GetIncidentsByService(ctx context.Context, serviceID string) ([]*Incident, error)
	GetRecentIncidents(ctx context.Context, limit int) ([]*Incident, error)
	GetActiveIncidents(ctx context.Context) ([]*Incident, error)

	// Service management
	SaveService(ctx context.Context, service *Service) error
	GetService(ctx context.Context, id string) (*Service, error)
	GetAllServices(ctx context.Context) ([]*Service, error)
	GetEnabledServices(ctx context.Context) ([]*Service, error)
	UpdateService(ctx context.Context, service *Service) error
	DeleteService(ctx context.Context, id string) error

	// Statistics
	GetServiceStats(ctx context.Context, serviceID string, since time.Time) (*ServiceStats, error)
	GetAllServicesIncidentStats(ctx context.Context) ([]*ServiceIncidentStats, error)

	// Cleanup
	Close() error
}

// Service represents a monitored service
type Service struct {
	ID        string        `json:"id" yaml:"id"`
	Name      string        `json:"name" yaml:"name"`
	Protocol  string        `json:"protocol" yaml:"protocol"`
	Endpoint  string        `json:"endpoint" yaml:"endpoint"`
	Interval  time.Duration `json:"interval" yaml:"interval" swaggertype:"primitive,integer"`
	Timeout   time.Duration `json:"timeout" yaml:"timeout" swaggertype:"primitive,integer"`
	Retries   int           `json:"retries" yaml:"retries"`
	Tags      []string      `json:"tags" yaml:"tags"`
	Config    MonitorConfig `json:"config" yaml:"config"`
	State     *ServiceState `json:"state,omitempty" yaml:"state,omitempty"`
	IsEnabled bool          `json:"is_enabled" yaml:"is_enabled"`
}

// ServiceState represents the current state of a monitored service
type ServiceState struct {
	Status             ServiceStatus `json:"status"`
	LastCheck          *time.Time    `json:"last_check,omitempty"`
	NextCheck          *time.Time    `json:"next_check,omitempty"`
	LastError          string        `json:"last_error,omitempty"`
	ConsecutiveFails   int           `json:"consecutive_fails"`
	ConsecutiveSuccess int           `json:"consecutive_success"`
	TotalChecks        int           `json:"total_checks"`
	ResponseTime       time.Duration `json:"response_time" swaggertype:"primitive,integer"`
}

// MarshalJSON кастомно сериализует LastCheck и NextCheck, чтобы если они нулевые — не попадали в json
func (s ServiceState) MarshalJSON() ([]byte, error) {
	type Alias ServiceState
	type outStruct struct {
		Alias
		LastCheck *time.Time `json:"last_check,omitempty"`
		NextCheck *time.Time `json:"next_check,omitempty"`
	}
	return json.Marshal(&outStruct{
		Alias:     (Alias)(s),
		LastCheck: s.LastCheck,
		NextCheck: s.NextCheck,
	})
}

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

// Incident represents a service incident
type Incident struct {
	ID        string         `json:"id"`
	ServiceID string         `json:"service_id"`
	StartTime time.Time      `json:"start_time"`
	EndTime   *time.Time     `json:"end_time,omitempty"`
	Error     string         `json:"error"`
	Duration  *time.Duration `json:"duration,omitempty" swaggertype:"primitive,integer"`
	Resolved  bool           `json:"resolved"`
}

// ServiceStats holds statistics for a service
type ServiceStats struct {
	ServiceID        string        `json:"service_id"`
	TotalIncidents   int           `json:"total_incidents"`
	TotalDowntime    time.Duration `json:"total_downtime" swaggertype:"primitive,integer"`
	UptimePercentage float64       `json:"uptime_percentage"`
	Period           time.Duration `json:"period" swaggertype:"primitive,integer"`
	AvgResponseTime  time.Duration `json:"avg_response_time" swaggertype:"primitive,integer"`
}

// ServiceIncidentStats holds incident statistics for a service
type ServiceIncidentStats struct {
	ServiceID       string `json:"service_id"`
	ActiveIncidents int    `json:"active_incidents"`
	TotalIncidents  int    `json:"total_incidents"`
}

// MonitorConfig represents configuration for different monitor types
type MonitorConfig struct {
	HTTP  *HTTPConfig  `json:"http,omitempty" yaml:"http,omitempty"`
	TCP   *TCPConfig   `json:"tcp,omitempty" yaml:"tcp,omitempty"`
	GRPC  *GRPCConfig  `json:"grpc,omitempty" yaml:"grpc,omitempty"`
	Redis *RedisConfig `json:"redis,omitempty" yaml:"redis,omitempty"`
}

// HTTPConfig represents HTTP/HTTPS monitor configuration
type HTTPConfig struct {
	Method         string            `json:"method" yaml:"method"`
	ExpectedStatus int               `json:"expected_status" yaml:"expected_status"`
	Headers        map[string]string `json:"headers" yaml:"headers"`
	ExtendedConfig map[string]any    `json:"extended_config,omitempty" yaml:"extended_config,omitempty"` // For multi-endpoint configuration
}

// TCPConfig represents TCP monitor configuration
type TCPConfig struct {
	SendData   string `json:"send_data" yaml:"send_data"`
	ExpectData string `json:"expect_data" yaml:"expect_data"`
}

// GRPCConfig represents gRPC monitor configuration
type GRPCConfig struct {
	CheckType   string `json:"check_type" yaml:"check_type"`
	ServiceName string `json:"service_name" yaml:"service_name"`
	TLS         bool   `json:"tls" yaml:"tls"`
	InsecureTLS bool   `json:"insecure_tls" yaml:"insecure_tls"`
}

// RedisConfig represents Redis monitor configuration
type RedisConfig struct {
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
}

// GenerateULID generates a new ULID
func GenerateULID() string {
	return ulid.Make().String()
}
