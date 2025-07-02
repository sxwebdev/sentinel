package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/cockroachdb/pebble/v2"
	"github.com/sxwebdev/chaindb"
	"github.com/sxwebdev/sentinel/internal/config"
)

// ChainDBStorage implements storage using ChainDB
type ChainDBStorage struct {
	db            chaindb.Database
	incidentTable chaindb.Table
}

var _ Storage = (*ChainDBStorage)(nil)

// NewChainDBStorage creates a new ChainDB storage instance
func NewChainDBStorage(dbPath string) (*ChainDBStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open ChainDB
	db, err := chaindb.NewDatabase(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open chaindb: %w", err)
	}

	// Create table for incidents
	incidentTable := chaindb.NewTable(db, []byte("incident:"))

	return &ChainDBStorage{
		db:            db,
		incidentTable: incidentTable,
	}, nil
}

// Close closes the database connection
func (s *ChainDBStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// SaveIncident saves a new incident to the database
func (s *ChainDBStorage) SaveIncident(ctx context.Context, incident *config.Incident) error {
	data, err := json.Marshal(incident)
	if err != nil {
		return fmt.Errorf("failed to marshal incident: %w", err)
	}

	key := fmt.Sprintf("%s:%s", incident.ServiceName, incident.ID)
	return s.incidentTable.Put([]byte(key), data)
}

// GetIncident retrieves an incident by ID
func (s *ChainDBStorage) GetIncident(ctx context.Context, serviceID, incidentID string) (*config.Incident, error) {
	key := fmt.Sprintf("%s:%s", serviceID, incidentID)
	data, err := s.incidentTable.Get([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("incident not found: %w", err)
	}

	var incident config.Incident
	if err := json.Unmarshal(data, &incident); err != nil {
		return nil, fmt.Errorf("failed to unmarshal incident: %w", err)
	}

	return &incident, nil
}

// UpdateIncident updates an existing incident
func (s *ChainDBStorage) UpdateIncident(ctx context.Context, incident *config.Incident) error {
	return s.SaveIncident(ctx, incident) // ChainDB handles updates the same way
}

// GetIncidentsByService retrieves all incidents for a specific service
func (s *ChainDBStorage) GetIncidentsByService(ctx context.Context, serviceName string) ([]*config.Incident, error) {
	prefix := []byte(fmt.Sprintf("%s:", serviceName))

	var incidents []*config.Incident

	// Create iterator with prefix bounds using Pebble IterOptions
	iterOptions := &pebble.IterOptions{
		LowerBound: prefix,
		UpperBound: append(prefix, 0xFF), // Next possible prefix
	}

	iter, err := s.incidentTable.NewIterator(ctx, iterOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Release()

	// Iterate over keys with the prefix
	for valid := iter.First(); valid && iter.Error() == nil; valid = iter.Next() {
		value := iter.Value()
		var incident config.Incident
		if err := json.Unmarshal(value, &incident); err != nil {
			return nil, fmt.Errorf("failed to unmarshal incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return incidents, nil
}

// GetRecentIncidents retrieves recent incidents across all services
func (s *ChainDBStorage) GetRecentIncidents(ctx context.Context, limit int) ([]*config.Incident, error) {
	var incidents []*config.Incident

	// Create iterator for the entire incident table
	iter, err := s.incidentTable.NewIterator(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Release()

	// Iterate over all incidents
	for valid := iter.First(); valid && iter.Error() == nil; valid = iter.Next() {
		value := iter.Value()
		var incident config.Incident
		if err := json.Unmarshal(value, &incident); err != nil {
			return nil, fmt.Errorf("failed to unmarshal incident: %w", err)
		}
		incidents = append(incidents, &incident)
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	// Sort by start time (most recent first) and limit
	if len(incidents) > 1 {
		for i := 0; i < len(incidents)-1; i++ {
			for j := i + 1; j < len(incidents); j++ {
				if incidents[i].StartTime.Before(incidents[j].StartTime) {
					incidents[i], incidents[j] = incidents[j], incidents[i]
				}
			}
		}
	}

	if limit > 0 && len(incidents) > limit {
		incidents = incidents[:limit]
	}

	return incidents, nil
}

// GetActiveIncidents retrieves all currently active (unresolved) incidents
func (s *ChainDBStorage) GetActiveIncidents(ctx context.Context) ([]*config.Incident, error) {
	var activeIncidents []*config.Incident

	// Create iterator for the incident table
	iter, err := s.incidentTable.NewIterator(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create iterator: %w", err)
	}
	defer iter.Release()

	// Iterate over all incidents
	for valid := iter.First(); valid && iter.Error() == nil; valid = iter.Next() {
		value := iter.Value()
		var incident config.Incident
		if err := json.Unmarshal(value, &incident); err != nil {
			return nil, fmt.Errorf("failed to unmarshal incident: %w", err)
		}

		if !incident.Resolved {
			activeIncidents = append(activeIncidents, &incident)
		}
	}

	if err := iter.Error(); err != nil {
		return nil, fmt.Errorf("iterator error: %w", err)
	}

	return activeIncidents, nil
}

// GetServiceStats calculates statistics for a service
func (s *ChainDBStorage) GetServiceStats(ctx context.Context, serviceName string, since time.Time) (*ServiceStats, error) {
	incidents, err := s.GetIncidentsByService(ctx, serviceName)
	if err != nil {
		return nil, err
	}

	stats := &ServiceStats{
		ServiceName:    serviceName,
		TotalIncidents: 0,
		TotalDowntime:  0,
		Period:         time.Since(since),
	}

	for _, incident := range incidents {
		if incident.StartTime.After(since) {
			stats.TotalIncidents++
			if incident.Duration != nil {
				stats.TotalDowntime += *incident.Duration
			} else if !incident.Resolved {
				// If incident is still active, calculate downtime from start time
				stats.TotalDowntime += time.Since(incident.StartTime)
			}
		}
	}

	// Calculate uptime percentage
	if stats.Period > 0 {
		uptimeRatio := float64(stats.Period-stats.TotalDowntime) / float64(stats.Period)
		stats.UptimePercentage = uptimeRatio * 100
		if stats.UptimePercentage < 0 {
			stats.UptimePercentage = 0
		}
	} else {
		stats.UptimePercentage = 100
	}

	return stats, nil
}

// ServiceStats holds statistics for a service
type ServiceStats struct {
	ServiceName      string        `json:"service_name"`
	TotalIncidents   int           `json:"total_incidents"`
	TotalDowntime    time.Duration `json:"total_downtime"`
	UptimePercentage float64       `json:"uptime_percentage"`
	Period           time.Duration `json:"period"`
}
