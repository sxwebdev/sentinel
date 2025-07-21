package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sxwebdev/sentinel/pkg/dbutils"
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

// Name returns the storage type
func (s *SQLiteStorage) Name() string {
	return "SQLite"
}

// Start initializes the storage
func (s *SQLiteStorage) Start(_ context.Context) error {
	if s.db == nil {
		return fmt.Errorf("storage not initialized")
	}
	return nil
}

// Stop closes the database connection
func (s *SQLiteStorage) Stop(_ context.Context) error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close database: %w", err)
		}
		s.db = nil
	}
	return nil
}

// Incident methods

// SaveIncident saves a new incident to the database
func (s *SQLiteStorage) SaveIncident(ctx context.Context, incident *Incident) error {
	incident.ID = GenerateULID()

	return s.orm.CreateIncident(ctx, incident)
}

// GetIncidentByID retrieves an incident by ID
func (s *SQLiteStorage) GetIncidentByID(ctx context.Context, id string) (*Incident, error) {
	return s.orm.GetIncidentByID(ctx, id)
}

// UpdateIncident updates an existing incident
func (s *SQLiteStorage) UpdateIncident(ctx context.Context, incident *Incident) error {
	return s.orm.UpdateIncident(ctx, incident)
}

// DeleteIncident deletes an incident by ID
func (s *SQLiteStorage) DeleteIncident(ctx context.Context, incidentID string) error {
	return s.orm.DeleteIncident(ctx, incidentID)
}

// FindIncidents retrieves all incidents
func (s *SQLiteStorage) FindIncidents(ctx context.Context, params FindIncidentsParams) (dbutils.FindResponseWithCount[*Incident], error) {
	return s.orm.FindIncidents(ctx, params)
}

// IncidentsCount retrieves the total count of incidents
func (s *SQLiteStorage) IncidentsCount(ctx context.Context, params FindIncidentsParams) (uint32, error) {
	return s.orm.IncidentsCount(ctx, params)
}

// GetServiceStats calculates statistics for a service
func (s *SQLiteStorage) GetServiceStats(ctx context.Context, params FindIncidentsParams) (*ServiceStats, error) {
	return s.orm.GetServiceStatsWithORM(ctx, params)
}

// ResolveAllIncidents resolves all incidents for a service
func (s *SQLiteStorage) ResolveAllIncidents(ctx context.Context, serviceID string) ([]*Incident, error) {
	return s.orm.ResolveAllIncidents(ctx, serviceID)
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
func (s *SQLiteStorage) FindServices(ctx context.Context, params FindServicesParams) (dbutils.FindResponseWithCount[*Service], error) {
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

// GetAllTags retrieves all unique tags across services
func (s *SQLiteStorage) GetAllTags(ctx context.Context) ([]string, error) {
	return s.orm.GetAllTags(ctx)
}

// GetAllTagsWithCount retrieves all unique tags with their usage count
func (s *SQLiteStorage) GetAllTagsWithCount(ctx context.Context) (map[string]int, error) {
	return s.orm.GetAllTagsWithCount(ctx)
}

// GetSQLiteVersion returns the SQLite version
func (s *SQLiteStorage) GetSQLiteVersion(ctx context.Context) (string, error) {
	var version string
	err := s.db.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("failed to get SQLite version: %w", err)
	}
	return version, nil
}
