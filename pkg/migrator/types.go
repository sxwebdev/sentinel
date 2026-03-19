package migrator

import (
	"context"
	"database/sql"
)

type logger interface {
	Infof(format string, args ...any)
	Errorf(format string, args ...any)
}

type migration struct {
	version int
	name    string
	upSQL   string
	downSQL string

	beforeUpFn   func(context.Context, *sql.Tx) error
	afterUpFn    func(context.Context, *sql.Tx) error
	beforeDownFn func(context.Context, *sql.Tx) error
	afterDownFn  func(context.Context, *sql.Tx) error
}

type DataMigration struct {
	Version int
	// Hooks
	BeforeUpFn   func(context.Context, *sql.Tx) error
	AfterUpFn    func(context.Context, *sql.Tx) error
	BeforeDownFn func(context.Context, *sql.Tx) error
	AfterDownFn  func(context.Context, *sql.Tx) error
}

type DataMigrations []DataMigration
