package sqlite

import (
	"context"
	"fmt"
)

// GetSQLiteVersion returns the SQLite version
func (o *SQLite) GetSQLiteVersion(ctx context.Context) (string, error) {
	var version string
	err := o.DB.QueryRowContext(ctx, "SELECT sqlite_version()").Scan(&version)
	if err != nil {
		return "", fmt.Errorf("failed to get SQLite version: %w", err)
	}
	return version, nil
}
