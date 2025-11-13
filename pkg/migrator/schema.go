package migrator

import "database/sql"

// schemaVersionTable creates the schema version tracking table
const schemaVersionTable = `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER PRIMARY KEY,
	applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

// GetCurrentSchemaVersion gets the current schema version from the database
func GetCurrentSchemaVersion(db *sql.DB) (int, error) {
	var version int
	err := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
