package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/monitor"
	"github.com/sxwebdev/sentinel/internal/notifier"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/scheduler"
	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/upgrader"
	"github.com/sxwebdev/sentinel/internal/web"
	"github.com/tkcrm/mx/launcher"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/service"
	"github.com/tkcrm/mx/service/pingpong"
	"github.com/urfave/cli/v3"
)

func startCMD() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "start the server",
		Flags: []cli.Flag{cfgPathsFlag()},
		Action: func(ctx context.Context, cl *cli.Command) error {
			conf, err := config.Load(cl.String("config"))
			if err != nil {
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

			// set default timezone
			time.Local, err = time.LoadLocation(conf.Timezone)
			if err != nil {
				return fmt.Errorf("failed to set timezone: %w", err)
			}

			// Initialize storage
			store, err := storage.NewStorage(storage.StorageTypeSQLite, conf.Database.Path)
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}

			// Print SQLite version if using SQLite storage
			sqliteVersion, err := store.GetSQLiteVersion(ctx)
			if err != nil {
				return fmt.Errorf("failed to get SQLite version: %w", err)
			}
			l.Infof("SQLite version: %s", sqliteVersion)

			// Initialize notifier
			var notif notifier.Notifier
			if conf.Notifications.Enabled {
				notif, err = notifier.NewNotifier(conf.Notifications.URLs)
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

			// Create monitor service
			monitorService := monitor.NewMonitorService(store, conf, notif, rc)

			// Initialize scheduler
			sched := scheduler.New(l, monitorService, rc)

			webServer, err := web.NewServer(l, conf, web.ServerInfo{
				Version:       version,
				CommitHash:    commitHash,
				BuildDate:     buildDate,
				GoVersion:     runtime.Version(),
				SqliteVersion: sqliteVersion,
				OS:            runtime.GOOS,
				Arch:          runtime.GOARCH,
			}, monitorService, store, rc, upgr)
			if err != nil {
				return fmt.Errorf("failed to initialize web server: %w", err)
			}

			// register services
			ln.ServicesRunner().Register(
				service.New(service.WithService(pingpong.New(l))),
				service.New(service.WithService(store)),
				service.New(service.WithService(rc)),
				service.New(service.WithService(sched)),
				service.New(service.WithService(webServer)),
			)

			return ln.Run()
		},
	}
}
