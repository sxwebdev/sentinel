package storage

import (
	"context"
	"time"

	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

// Storage defines the interface for incident storage
type Storage interface {
	// Incident management
	SaveIncident(ctx context.Context, incident *Incident) error
	UpdateIncident(ctx context.Context, incident *Incident) error
	DeleteIncident(ctx context.Context, incidentID string) error
	GetIncidentsByService(ctx context.Context, serviceID string) ([]*Incident, error)
	GetRecentIncidents(ctx context.Context, limit int) ([]*Incident, error)
	GetActiveIncidents(ctx context.Context) ([]*Incident, error)

	// Service management
	CreateService(ctx context.Context, request CreateUpdateServiceRequest) (*Service, error)
	GetServiceByID(ctx context.Context, id string) (*Service, error)
	FindServices(ctx context.Context, params FindServicesParams) (dbutils.FindResponseWithCount[*Service], error)
	UpdateService(ctx context.Context, id string, request CreateUpdateServiceRequest) (*Service, error)
	DeleteService(ctx context.Context, id string) error

	// Service state management
	GetServiceState(ctx context.Context, serviceID string) (*ServiceStateRecord, error)
	UpdateServiceState(ctx context.Context, state *ServiceStateRecord) error
	GetAllServiceStates(ctx context.Context) ([]*ServiceStateRecord, error)

	// Tags
	GetAllTags(ctx context.Context) ([]string, error)
	GetAllTagsWithCount(ctx context.Context) (map[string]int, error)

	// Statistics
	GetServiceStats(ctx context.Context, serviceID string, since time.Time) (*ServiceStats, error)

	// Cleanup
	Close() error
}
