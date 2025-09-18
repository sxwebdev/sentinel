package migrations

import (
	"database/sql"
	"fmt"
)

type applyMigrationType int

const (
	applyMigrationTypeUp applyMigrationType = iota
	applyMigrationTypeDown
)

// string returns the string representation of the applyMigrationType
func (at applyMigrationType) String() string {
	switch at {
	case applyMigrationTypeUp:
		return "up"
	case applyMigrationTypeDown:
		return "down"
	default:
		return "unknown"
	}
}

// applyMigration runs a single migration
func (m *Migrations) applyMigration(db *sql.DB, at applyMigrationType, version int, sql string) error {
	if sql == "" {
		return nil
	}

	m.info("applying migration %s version %d", at.String(), version)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Execute migration SQL
	if _, err := tx.Exec(sql); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Record migration version
	switch at {
	case applyMigrationTypeUp:
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration version: %w", err)
		}
	case applyMigrationTypeDown:
		if _, err := tx.Exec("DELETE FROM schema_version WHERE version = ?", version); err != nil {
			return fmt.Errorf("failed to remove migration version record: %w", err)
		}
	default:
		return fmt.Errorf("unknown apply type: %d", at)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	return nil
}
