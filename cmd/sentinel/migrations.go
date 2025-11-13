package main

import (
	"github.com/sxwebdev/sentinel/internal/datamigrations"
	"github.com/sxwebdev/sentinel/pkg/migrator"
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
	return migrator.CliCmd(l, sql.MigrationsFS, sql.MigrationsPath, datamigrations.Migrations)
}
