package storage

import (
	"database/sql"
	"fmt"
)

// Migration represents a database migration
type Migration struct {
	Version int
	SQL     string
}

// migrations contains all database migrations in order
var migrations = []Migration{
	{
		Version: 1,
		SQL: `
		-- Create services table
		CREATE TABLE IF NOT EXISTS services (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			protocol TEXT NOT NULL,
			endpoint TEXT NOT NULL,
			interval TEXT NOT NULL,
			timeout TEXT NOT NULL,
			retries INTEGER NOT NULL DEFAULT 3,
			tags jsonb NOT NULL DEFAULT '[]',
			config jsonb NOT NULL DEFAULT '{}',
			state jsonb NOT NULL DEFAULT '{}',
			is_active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Create incidents table
		CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			service_id TEXT NOT NULL REFERENCES services(id),
			start_time DATETIME NOT NULL,
			end_time DATETIME,
			error TEXT NOT NULL,
			duration_ns INTEGER,
			resolved BOOLEAN NOT NULL DEFAULT FALSE,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);

		-- Create indexes for performance
		CREATE INDEX IF NOT EXISTS idx_services_name ON services(name);
		CREATE INDEX IF NOT EXISTS idx_services_protocol ON services(protocol);
		CREATE INDEX IF NOT EXISTS idx_incidents_service_id ON incidents(service_id);
		CREATE INDEX IF NOT EXISTS idx_incidents_start_time ON incidents(start_time);
		CREATE INDEX IF NOT EXISTS idx_incidents_resolved ON incidents(resolved);
		CREATE INDEX IF NOT EXISTS idx_incidents_service_resolved ON incidents(service_id, resolved);
		`,
	},
}

// schemaVersionTable creates the schema version tracking table
const schemaVersionTable = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER PRIMARY KEY,
	applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// runMigrations runs all pending database migrations
func runMigrations(db *sql.DB) error {
	// Create schema version table if it doesn't exist
	if _, err := db.Exec(schemaVersionTable); err != nil {
		return fmt.Errorf("failed to create schema version table: %w", err)
	}

	// Get current schema version
	currentVersion, err := getCurrentSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := runMigration(db, migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}
		}
	}

	return nil
}

// getCurrentSchemaVersion gets the current schema version from the database
func getCurrentSchemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}

// runMigration runs a single migration
func runMigration(db *sql.DB, migration Migration) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(migration.SQL); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration version
	if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", migration.Version); err != nil {
		return fmt.Errorf("failed to record migration version: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}
