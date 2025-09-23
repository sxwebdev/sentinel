package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"connectrpc.com/connect"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/datamigrations"
	"github.com/sxwebdev/sentinel/internal/handlerutils"
	"github.com/sxwebdev/sentinel/internal/hubserver"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/scheduler"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/upgrader"
	"github.com/sxwebdev/sentinel/internal/web"
	"github.com/sxwebdev/sentinel/pkg/migrations"
	"github.com/sxwebdev/sentinel/pkg/sqlite"
	"github.com/sxwebdev/sentinel/sql"
	"github.com/tkcrm/mx/launcher"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/service"
	"github.com/tkcrm/mx/service/pingpong"
	"github.com/tkcrm/mx/transport/connectrpc_transport"
	"github.com/urfave/cli/v3"
	"go.akshayshah.org/connectproto"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/protobuf/encoding/protojson"
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

			dbPath := filepath.Join(conf.DataDir, "sqlite", sqliteDBFile)

			// init sqlite
			db, err := sqlite.New(ctx, dbPath)
			if err != nil {
				return fmt.Errorf("failed to initialize sqlite: %w", err)
			}

			// Print SQLite version if using SQLite storage
			sqliteVersion, err := db.GetSQLiteVersion(ctx)
			if err != nil {
				return fmt.Errorf("failed to get SQLite version: %w", err)
			}
			l.Infof("SQLite version: %s", sqliteVersion)

			// check and run all migrations
			m := migrations.New(l, sql.MigrationsFS, sql.MigrationsPath, datamigrations.Migrations)
			if err := m.MigrateUpAll(ctx, dbPath); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}

			st, err := store.New(db.DB)
			if err != nil {
				return fmt.Errorf("failed to initialize store: %w", err)
			}

			// Init receiver
			rc := receiver.New()

			// Initialize upgrader if configured
			upgr, err := upgrader.New(l, conf.Upgrader)
			if err != nil {
				return fmt.Errorf("failed to initialize upgrader: %w", err)
			}

			baseServices := baseservices.New(l, st, rc)

			// Initialize scheduler
			sched := scheduler.New(l, rc, baseServices)

			serverInfo := models.GetSystemInfo(version, commitHash, buildDate)
			serverInfo.SqliteVersion = sqliteVersion

			webServer, err := web.NewServer(l, conf, serverInfo, baseServices, rc, upgr)
			if err != nil {
				return fmt.Errorf("failed to initialize web server: %w", err)
			}

			// init agent rpc server
			hubServer := hubserver.New(baseServices)

			rpcServer := connectrpc_transport.NewServer(
				connectrpc_transport.WithName("hub-server"),
				connectrpc_transport.WithLogger(l),
				connectrpc_transport.WithConfig(conf.HubServer),
				connectrpc_transport.WithServices(hubServer),
				connectrpc_transport.WithServerHandlerWrapper(
					func(h http.Handler) http.Handler {
						return h2c.NewHandler(
							handlerutils.WithCORS(h),
							&http2.Server{})
					},
				),
				connectrpc_transport.WithReflection(
					hubServer.Name(),
				),
				connectrpc_transport.WithConnectRPCOptions(
					connect.WithHandlerOptions(
						connectproto.WithJSON(
							protojson.MarshalOptions{EmitUnpopulated: true},
							protojson.UnmarshalOptions{DiscardUnknown: true},
						),
					),
				),
			)

			// register services
			ln.ServicesRunner().Register(
				service.New(service.WithService(pingpong.New(l))),
				service.New(service.WithService(db)),
				service.New(service.WithService(rc)),
				service.New(service.WithService(sched)),
				service.New(service.WithService(webServer)),
				service.New(service.WithService(rpcServer)),
				service.New(service.WithService(baseServices.Notifications().Sender())),
			)

			return ln.Run()
		},
	}
}
