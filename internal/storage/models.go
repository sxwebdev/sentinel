package storage

import (
	"context"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
)

// Storage defines the interface for incident storage
type Storage interface {
	// Incident management
	SaveIncident(ctx context.Context, incident *config.Incident) error
	GetIncident(ctx context.Context, serviceID, incidentID string) (*config.Incident, error)
	UpdateIncident(ctx context.Context, incident *config.Incident) error
	GetIncidentsByService(ctx context.Context, serviceName string) ([]*config.Incident, error)
	GetRecentIncidents(ctx context.Context, limit int) ([]*config.Incident, error)
	GetActiveIncidents(ctx context.Context) ([]*config.Incident, error)

	// Statistics
	GetServiceStats(ctx context.Context, serviceName string, since time.Time) (*ServiceStats, error)

	// Cleanup
	Close() error
}
