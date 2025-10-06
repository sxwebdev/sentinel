package badgerdb

import (
	"context"

	"github.com/dgraph-io/badger/v4"
	"github.com/tkcrm/mx/logger"
)

type DB struct {
	db *badger.DB
}

func New(l logger.Logger, path string) (*DB, error) {
	defaultOptions := badger.DefaultOptions(path)
	defaultOptions.Logger = &bLogger{l: logger.With(l, "service", "badgerdb")}

	db, err := badger.Open(defaultOptions)
	if err != nil {
		return nil, err
	}
	return &DB{db: db}, nil
}

// Named returns the name of the database engine.
func (d *DB) Name() string {
	return "badgerdb"
}

// Start starts the database engine.
func (d *DB) Start(_ context.Context) error {
	return nil
}

// Stop stops the database engine.
func (d *DB) Stop(_ context.Context) error {
	return d.db.Close()
}
