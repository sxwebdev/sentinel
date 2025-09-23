package migrations

import (
	"database/sql"
	"fmt"

	"github.com/sxwebdev/sentinel/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func initDatabase(dbPath string) (*sql.DB, error) {
	dsn := sqlite.GetDSN(dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}
