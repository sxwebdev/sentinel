package migrator

import (
	"context"
	"fmt"
)

// MigrateUpAll runs all pending database migrations
func (m *Service) MigrateUpAll(ctx context.Context, dbPath string) error {
	m.info("applying all migrations")

	migrations, err := m.load()
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
		if migration.version > currentVersion {
			if err := m.applyMigration(ctx, db, applyMigrationTypeUp, migration); err != nil {
				return fmt.Errorf("failed to run migration %d: %w", migration.version, err)
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
