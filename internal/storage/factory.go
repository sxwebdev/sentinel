package storage

import (
	"fmt"
	"strings"
)

// StorageType represents the type of storage backend
type StorageType string

const (
	StorageTypeSQLite StorageType = "sqlite"
)

// NewStorage creates a new storage instance based on the provided type and path
func NewStorage(storageType StorageType, dbPath string) (Storage, error) {
	switch strings.ToLower(string(storageType)) {
	case string(StorageTypeSQLite):
		return NewSQLiteStorage(dbPath)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", storageType)
	}
}
