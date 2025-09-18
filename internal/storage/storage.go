package storage

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sxwebdev/sentinel/pkg/migrations"
	"github.com/tkcrm/mx/logger"
	_ "modernc.org/sqlite"

	emsql "github.com/sxwebdev/sentinel/sql"
)

// Storage implements Storage interface using SQLite
type Storage struct {
	db *sql.DB
}

// New creates a new SQLite storage instance
func New(l logger.Logger, dbPath string) (*Storage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	// Run migrations
	m := migrations.New(l, emsql.MigrationsFS, emsql.MigrationsPath)
	if err := m.MigrateUpAll(dbPath); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
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

	return &Storage{
		db: db,
	}, nil
}

// Name returns the storage type
func (s *Storage) Name() string {
	return "store"
}

// Start initializes the storage
func (s *Storage) Start(_ context.Context) error {
	if s.db == nil {
		return fmt.Errorf("sqlite not initialized")
	}
	return nil
}

// Stop closes the database connection
func (s *Storage) Stop(_ context.Context) error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("failed to close sqlite database: %w", err)
		}
		s.db = nil
	}
	return nil
}

// SQLiteDB returns the underlying sql.DB instance
func (s *Storage) SQLiteDB() *sql.DB {
	return s.db
}
