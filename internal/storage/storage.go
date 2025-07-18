package storage

import (
	"context"

	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

// Storage defines the interface for incident storage
type Storage interface {
	// Incident management
	GetIncidentByID(ctx context.Context, id string) (*Incident, error)
	SaveIncident(ctx context.Context, incident *Incident) error
	UpdateIncident(ctx context.Context, incident *Incident) error
	DeleteIncident(ctx context.Context, incidentID string) error
	FindIncidents(ctx context.Context, params FindIncidentsParams) (dbutils.FindResponseWithCount[*Incident], error)
	IncidentsCount(ctx context.Context, params FindIncidentsParams) (uint32, error)
	ResolveAllIncidents(ctx context.Context, serviceID string) ([]*Incident, error)

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
	GetServiceStats(ctx context.Context, params FindIncidentsParams) (*ServiceStats, error)

	// Cleanup
	Close() error
}
