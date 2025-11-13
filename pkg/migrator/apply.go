package migrator

import (
	"context"
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
func (m *Service) applyMigration(ctx context.Context, db *sql.DB, at applyMigrationType, migration migration) error {
	sql := migration.upSQL
	if at == applyMigrationTypeDown {
		sql = migration.downSQL
	}

	m.info("applying migration %s version %d", at.String(), migration.version)

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Run Before hook
	if at == applyMigrationTypeUp && migration.beforeUpFn != nil {
		if err := migration.beforeUpFn(ctx, tx); err != nil {
			return fmt.Errorf("before up hook failed: %w", err)
		}
	} else if at == applyMigrationTypeDown && migration.beforeDownFn != nil {
		if err := migration.beforeDownFn(ctx, tx); err != nil {
			return fmt.Errorf("before down hook failed: %w", err)
		}
	}

	// Execute migration SQL
	if _, err := tx.Exec(sql); err != nil {
		return fmt.Errorf("failed to execute migration SQL: %w", err)
	}

	// Run After hook
	if at == applyMigrationTypeUp && migration.afterUpFn != nil {
		if err := migration.afterUpFn(ctx, tx); err != nil {
			return fmt.Errorf("after up hook failed: %w", err)
		}
	} else if at == applyMigrationTypeDown && migration.afterDownFn != nil {
		if err := migration.afterDownFn(ctx, tx); err != nil {
			return fmt.Errorf("after down hook failed: %w", err)
		}
	}

	// Record migration version
	switch at {
	case applyMigrationTypeUp:
		if _, err := tx.Exec("INSERT INTO schema_version (version) VALUES (?)", migration.version); err != nil {
			return fmt.Errorf("failed to record migration version: %w", err)
		}
	case applyMigrationTypeDown:
		if _, err := tx.Exec("DELETE FROM schema_version WHERE version = ?", migration.version); err != nil {
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
