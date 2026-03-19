package migrator

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/sxwebdev/sentinel/pkg/sqlite"
	_ "modernc.org/sqlite"
)

func initDatabase(dbPath string) (*sql.DB, error) {
	// check if db file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("database file does not exist: %w", err)
	}

	dsn := sqlite.GetDSN(dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	return db, nil
}
