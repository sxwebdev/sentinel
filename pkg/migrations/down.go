package migrations

import (
	"fmt"
)

// MigrateDown rolls back the last applied migration
func (m *Migrations) MigrateDown(dbPath string) error {
	migrations, err := m.loadFromFS()
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
		if migrations[i].Version == currentVersion {
			migrationToRollback = &migrations[i]
			break
		}
	}

	if migrationToRollback == nil {
		return fmt.Errorf("no migration found to roll back for version %d", currentVersion)
	}

	// Run the rollback
	if err := m.applyMigration(db, applyMigrationTypeDown, migrationToRollback.Version, migrationToRollback.DownSQL); err != nil {
		return fmt.Errorf("failed to roll back migration %d: %w", migrationToRollback.Version, err)
	}

	return nil
}
