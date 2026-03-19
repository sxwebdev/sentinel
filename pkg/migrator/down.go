package migrator

import (
	"context"
	"fmt"
)

// MigrateDown rolls back the last applied migration
func (m *Service) MigrateDown(ctx context.Context, dbPath string) error {
	migrations, err := m.load()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	db, err := initDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Get current schema version
	currentVersion, err := GetCurrentSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	// Find the migration to roll back
	var migrationToRollback *migration
	for i := len(migrations) - 1; i >= 0; i-- {
		if migrations[i].version == currentVersion {
			migrationToRollback = &migrations[i]
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("no migration found to roll back for version %d", currentVersion)
	}

	// Run the rollback
	if err := m.applyMigration(ctx, db, applyMigrationTypeDown, *migrationToRollback); err != nil {
		return fmt.Errorf("failed to roll back migration %d: %w", migrationToRollback.version, err)
	}

	return nil
}
