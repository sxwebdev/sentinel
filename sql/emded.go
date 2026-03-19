package sql

import (
	"embed"
)

//go:embed migrations/*.sql
var MigrationsFS embed.FS

var MigrationsPath = "migrations"
