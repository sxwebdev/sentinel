package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	_ "modernc.org/sqlite"
)

// IncidentRow represents a database row for incidents
type IncidentRow struct {
	ID          string     `db:"id"`
	ServiceName string     `db:"service_name"`
	StartTime   time.Time  `db:"start_time"`
	EndTime     *time.Time `db:"end_time"`
	Error       string     `db:"error"`
	DurationNS  *int64     `db:"duration_ns"`
	Resolved    bool       `db:"resolved"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

// SQLiteStorage implements storage using SQLite
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

	// Open SQLite database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	storage := &SQLiteStorage{
		db:  db,
		orm: NewORMStorage(db),
	}

	// Initialize database schema
	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
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

// SaveIncident saves a new incident to the database
func (s *SQLiteStorage) SaveIncident(ctx context.Context, incident *config.Incident) error {
	return s.orm.CreateIncident(ctx, incident)
}

// GetIncident retrieves an incident by ID
func (s *SQLiteStorage) GetIncident(ctx context.Context, serviceID, incidentID string) (*config.Incident, error) {
	return s.orm.FindIncidentByID(ctx, serviceID, incidentID)
}

// UpdateIncident updates an existing incident
func (s *SQLiteStorage) UpdateIncident(ctx context.Context, incident *config.Incident) error {
	return s.orm.UpdateIncident(ctx, incident)
}

// GetIncidentsByService retrieves all incidents for a specific service
func (s *SQLiteStorage) GetIncidentsByService(ctx context.Context, serviceName string) ([]*config.Incident, error) {
	return s.orm.FindIncidentsByService(ctx, serviceName)
}

// GetRecentIncidents retrieves recent incidents across all services
func (s *SQLiteStorage) GetRecentIncidents(ctx context.Context, limit int) ([]*config.Incident, error) {
	return s.orm.FindRecentIncidents(ctx, limit)
}

// GetActiveIncidents retrieves all currently active (unresolved) incidents
func (s *SQLiteStorage) GetActiveIncidents(ctx context.Context) ([]*config.Incident, error) {
	return s.orm.FindActiveIncidents(ctx)
}

// GetServiceStats calculates statistics for a service
func (s *SQLiteStorage) GetServiceStats(ctx context.Context, serviceName string, since time.Time) (*ServiceStats, error) {
	return s.orm.GetServiceStatsWithORM(ctx, serviceName, since)
}
