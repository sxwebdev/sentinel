package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStorage implements Storage interface using SQLite
type SQLiteStorage struct {
	db  *sql.DB
	orm *ORMStorage
}

var _ Storage = (*SQLiteStorage)(nil)

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Open SQLite database with proper settings for concurrent access
	db, err := sql.Open("sqlite", dbPath+"?_busy_timeout=30000&_journal_mode=WAL&_synchronous=NORMAL&_cache_size=10000&_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(time.Hour)

	// Test connection
	if err := db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{
		db:  db,
		orm: NewORMStorage(db),
	}

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// Incident methods

// SaveIncident saves a new incident to the database
func (s *SQLiteStorage) SaveIncident(ctx context.Context, incident *Incident) error {
	incident.ID = GenerateULID()

	return s.orm.CreateIncident(ctx, incident)
}

// UpdateIncident updates an existing incident
func (s *SQLiteStorage) UpdateIncident(ctx context.Context, incident *Incident) error {
	return s.orm.UpdateIncident(ctx, incident)
}

// DeleteIncident deletes an incident by ID
func (s *SQLiteStorage) DeleteIncident(ctx context.Context, incidentID string) error {
	return s.orm.DeleteIncident(ctx, incidentID)
}

// GetIncidentsByService retrieves all incidents for a specific service
func (s *SQLiteStorage) GetIncidentsByService(ctx context.Context, serviceID string) ([]*Incident, error) {
	return s.orm.FindIncidentsByService(ctx, serviceID)
}

// GetRecentIncidents retrieves recent incidents across all services
func (s *SQLiteStorage) GetRecentIncidents(ctx context.Context, limit int) ([]*Incident, error) {
	return s.orm.FindRecentIncidents(ctx, limit)
}

// GetActiveIncidents retrieves all currently active (unresolved) incidents
func (s *SQLiteStorage) GetActiveIncidents(ctx context.Context) ([]*Incident, error) {
	return s.orm.FindActiveIncidents(ctx)
}

// GetServiceStats calculates statistics for a service
func (s *SQLiteStorage) GetServiceStats(ctx context.Context, serviceID string, since time.Time) (*ServiceStats, error) {
	return s.orm.GetServiceStatsWithORM(ctx, serviceID, since)
}

// Service methods

// CreateService saves a new service to the database
func (s *SQLiteStorage) CreateService(ctx context.Context, service CreateUpdateServiceRequest) (*Service, error) {
	return s.orm.CreateService(ctx, service)
}

// GetService retrieves a service by ID
func (s *SQLiteStorage) GetServiceByID(ctx context.Context, id string) (*Service, error) {
	return s.orm.GetServiceByID(ctx, id)
}

// FindServices retrieves all services
func (s *SQLiteStorage) FindServices(ctx context.Context, params FindServicesParams) ([]*Service, error) {
	return s.orm.FindServices(ctx, params)
}

// UpdateService updates an existing service
func (s *SQLiteStorage) UpdateService(ctx context.Context, id string, service CreateUpdateServiceRequest) (*Service, error) {
	return s.orm.UpdateService(ctx, id, service)
}

// DeleteService deletes a service by ID
func (s *SQLiteStorage) DeleteService(ctx context.Context, id string) error {
	return s.orm.DeleteService(ctx, id)
}

// Service state management methods

// GetServiceState gets service state
func (s *SQLiteStorage) GetServiceState(ctx context.Context, serviceID string) (*ServiceStateRecord, error) {
	return s.orm.GetServiceState(ctx, serviceID)
}

// UpdateServiceState updates service state
func (s *SQLiteStorage) UpdateServiceState(ctx context.Context, state *ServiceStateRecord) error {
	return s.orm.UpdateServiceState(ctx, state)
}

// GetAllServiceStates gets all service states
func (s *SQLiteStorage) GetAllServiceStates(ctx context.Context) ([]*ServiceStateRecord, error) {
	return s.orm.GetAllServiceStates(ctx)
}
