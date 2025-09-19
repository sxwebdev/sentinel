package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/monitor"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/scheduler"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/upgrader"
	"github.com/sxwebdev/sentinel/internal/web"
	"github.com/sxwebdev/sentinel/pkg/sqlite"
	"github.com/tkcrm/mx/launcher"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/service"
	"github.com/tkcrm/mx/service/pingpong"
	"github.com/urfave/cli/v3"
)

func hubStartCMD() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "start the server",
		Flags: []cli.Flag{cfgPathsFlag()},
		Action: func(ctx context.Context, cl *cli.Command) error {
			conf := new(config.ConfigHub)
			if err := config.Load(conf, envHubPrefix, cl.StringSlice("config")); err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			loggerOpts := append(defaultLoggerOpts(), logger.WithConfig(conf.Log))

			l := logger.NewExtended(loggerOpts...)
			defer func() {
				_ = l.Sync()
			}()

			// init launcher
			ln := launcher.New(
				launcher.WithVersion(version),
				launcher.WithName(appName),
				launcher.WithLogger(l),
				launcher.WithContext(ctx),
				launcher.WithRunnerServicesSequence(launcher.RunnerServicesSequenceFifo),
				launcher.WithOpsConfig(conf.Ops),
				launcher.WithAppStartStopLog(true),
			)

			// check if exists data dir, if not create it
			if _, err := os.Stat(conf.DataDir); os.IsNotExist(err) {
				if err := os.MkdirAll(conf.DataDir, 0o700); err != nil {
					return fmt.Errorf("failed to create data dir: %w", err)
				}
			}

			// set default timezone
			var err error
			time.Local, err = time.LoadLocation(conf.Timezone)
			if err != nil {
				return fmt.Errorf("failed to set timezone: %w", err)
			}

			dbPath := filepath.Join(conf.DataDir, sqliteDBFile)

			// init sqlite
			db, err := sqlite.New(ctx, dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize sqlite: %w", err)
			}

			st, err := store.New(db.DB)
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}

			// Print SQLite version if using SQLite storage
			sqliteVersion, err := db.GetSQLiteVersion(ctx)
			if err != nil {
				return fmt.Errorf("failed to get SQLite version: %w", err)
			}
			l.Infof("SQLite version: %s", sqliteVersion)

			// Initialize notifier
			var notif *notifier.Notifier
			if conf.Notifications.Enabled {
				notif, err = notifier.New(l, conf.Notifications.URLs)
				if err != nil {
					return fmt.Errorf("failed to initialize notifier: %w", err)
				}
			}

			// Init receiver
			rc := receiver.New()

			// Initialize upgrader if configured
			upgr, err := upgrader.New(l, conf.Upgrader)
			if err != nil {
				return fmt.Errorf("failed to initialize upgrader: %w", err)
			}

			baseServices := baseservices.New(st, rc)

			// Create monitor service
			monitorService := monitor.NewMonitorService(l, st, notif, baseServices)

			// Initialize scheduler
			sched := scheduler.New(l, monitorService, rc, baseServices)

			serverInfo := models.GetSystemInfo(version, commitHash, buildDate)
			serverInfo.SqliteVersion = sqliteVersion

			webServer, err := web.NewServer(l, conf, serverInfo, baseServices, monitorService, rc, upgr)
			if err != nil {
				return fmt.Errorf("failed to initialize web server: %w", err)
			}

			// register services
			ln.ServicesRunner().Register(
				service.New(service.WithService(pingpong.New(l))),
				service.New(service.WithService(db)),
				service.New(service.WithService(rc)),
				service.New(service.WithService(sched)),
				service.New(service.WithService(webServer)),
			)

			if notif != nil {
				ln.ServicesRunner().Register(
					service.New(service.WithService(notif)),
				)
			}

			return ln.Run()
		},
	}
}
