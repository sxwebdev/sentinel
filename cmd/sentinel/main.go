package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"

	mxsignal "github.com/tkcrm/mx/util/signal"
	"github.com/urfave/cli/v3"

	"github.com/tkcrm/mx/logger"
)

var (
	appName    = "sentinel"
	version    = "local"
	commitHash = "unknown"
	buildDate  = "unknown"
)

func getBuildVersion() string {
	return fmt.Sprintf(
		"\nrelease: %s\ncommit hash: %s\nbuild date: %s\ngo version: %s",
		version,
		commitHash,
		buildDate,
		runtime.Version(),
	)
}

func defaultLoggerOpts() []logger.Option {
	return []logger.Option{
		logger.WithAppName(appName),
		logger.WithAppVersion(version),
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), mxsignal.Shutdown()...)
	defer cancel()

	l := logger.NewExtended(defaultLoggerOpts()...)

	app := &cli.Command{
		Name:    appName,
		Usage:   "A CLI application for " + appName,
		Version: getBuildVersion(),
		Suggest: true,
		Commands: []*cli.Command{
			startCMD(),
			configCMD(),
			versionCMD(),
		},
	}

	// run cli runner
	if err := app.Run(ctx, os.Args); err != nil {
		l.Fatalf("failed to run cli runner: %s", err)
	}
}
