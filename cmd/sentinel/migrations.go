package main

import (
	"github.com/sxwebdev/sentinel/internal/datamigrations"
	"github.com/sxwebdev/sentinel/pkg/migrations"
	"github.com/sxwebdev/sentinel/sql"
	"github.com/tkcrm/mx/logger"
	"github.com/urfave/cli/v3"
)

func migrationsCMD() *cli.Command {
	opts := append(
		defaultLoggerOpts(),
		logger.WithConsoleColored(true),
		logger.WithLogFormat(logger.LoggerFormatConsole),
	)
	l := logger.NewExtended(opts...)
	return migrations.CliCmd(l, sql.MigrationsFS, sql.MigrationsPath, datamigrations.Migrations)
}
