package migrations

import (
	"fmt"
)

// MigrateUpAll runs all pending database migrations
func (m *Migrations) MigrateUpAll(dbPath string) error {
	m.info("run all migrations")

	migrations, err := m.loadFromFS()
	if err != nil {
		return fmt.Errorf("failed to load migrations: %w", err)
	}

	db, err := initDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %w", err)
	}
	defer db.Close()

	// Create schema version table if it doesn't exist
	if _, err := db.Exec(schemaVersionTable); err != nil {
		return fmt.Errorf("failed to create schema version table: %w", err)
	}

	// Get current schema version
	currentVersion, err := GetCurrentSchemaVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get current schema version: %w", err)
	}

	m.info("current schema version: %d", currentVersion)

	var appliedMigrationsCount int

	// Run pending migrations
	for _, migration := range migrations {
		if migration.Version > currentVersion {
			if err := m.applyMigration(db, applyMigrationTypeUp, migration.Version, migration.UpSQL); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.Version, err)
			}
			appliedMigrationsCount++
		}
	}

	if appliedMigrationsCount == 0 {
		m.info("no new migrations to apply")
	} else {
		m.info("applied %d new migrations", appliedMigrationsCount)
	}

	return nil
}
