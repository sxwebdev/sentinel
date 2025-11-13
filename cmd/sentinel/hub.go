package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sxwebdev/sentinel/internal/alertresolver"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/datamigrations"
	"github.com/sxwebdev/sentinel/internal/dispatcher"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/scheduler"
	"github.com/sxwebdev/sentinel/internal/servers"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/store/badgerdb"
	updater "github.com/sxwebdev/sentinel/internal/updated"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/sxwebdev/sentinel/pkg/locker"
	"github.com/sxwebdev/sentinel/pkg/migrator"
	"github.com/sxwebdev/sentinel/pkg/sqlite"
	"github.com/sxwebdev/sentinel/sql"
	"github.com/sxwebdev/tokenmanager"
	"github.com/tkcrm/mx/launcher"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/service"
	"github.com/tkcrm/mx/service/pingpong"
	"github.com/urfave/cli/v3"
)

func hubStartCMD() *cli.Command {
	return &cli.Command{
		Name:  "hub",
		Usage: "hub commands",
		Commands: []*cli.Command{
			{
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
						launcher.WithRunnerServicesSequence(launcher.RunnerServicesSequenceLifo),
						launcher.WithOpsConfig(conf.Ops),
						launcher.WithAppStartStopLog(true),
					)

					// check if exists data dir, if not create it
					if _, err := os.Stat(conf.HubDataDir()); os.IsNotExist(err) {
						l.Infof("creating data directory in %s", conf.HubDataDir())
						if err := os.MkdirAll(conf.HubDataDir(), 0o700); err != nil {
							return fmt.Errorf("failed to create data dir: %w", err)
						}
					}

					var authConfig config.AuthConfig

					// get file datadir/hub/secrets.json
					// if not exists create it with generated values
					secretsFilePath := filepath.Join(conf.HubDataDir(), "secrets.json")
					if _, err := os.Stat(secretsFilePath); os.IsNotExist(err) {
						l.Infof("creating secrets file in %s", secretsFilePath)
						if err := os.MkdirAll(filepath.Dir(secretsFilePath), 0o700); err != nil {
							return fmt.Errorf("failed to create secrets dir: %w", err)
						}

						accessToken, err := utils.GenerateRandomString(48, "")
						if err != nil {
							return fmt.Errorf("failed to generate access token secret key: %w", err)
						}

						refreshToken, err := utils.GenerateRandomString(48, "")
						if err != nil {
							return fmt.Errorf("failed to generate refresh token secret key: %w", err)
						}

						authConfig = config.AuthConfig{
							AccessTokenSecretKey:  accessToken,
							RefreshTokenSecretKey: refreshToken,
						}

						data, err := json.MarshalIndent(authConfig, "", "  ")
						if err != nil {
							return fmt.Errorf("failed to marshal secrets: %w", err)
						}

						if err := os.WriteFile(secretsFilePath, data, 0o600); err != nil {
							return fmt.Errorf("failed to create secrets file: %w", err)
						}
					} else {
						data, err := os.ReadFile(secretsFilePath)
						if err != nil {
							return fmt.Errorf("failed to read secrets file: %w", err)
						}

						if err := json.Unmarshal(data, &authConfig); err != nil {
							return fmt.Errorf("failed to unmarshal secrets file: %w", err)
						}

						if authConfig.AccessTokenSecretKey == "" || authConfig.RefreshTokenSecretKey == "" {
							return fmt.Errorf("invalid secrets file: missing keys")
						}
					}

					// set default timezone
					var err error
					time.Local, err = time.LoadLocation(conf.Timezone)
					if err != nil {
						return fmt.Errorf("failed to set timezone: %w", err)
					}

					sqliteDbPath := filepath.Join(conf.HubDataDir(), "sqlite", sqliteDBFile)

					// init sqlite
					sqliteDB, err := sqlite.New(ctx, sqliteDbPath)
					if err != nil {
						return fmt.Errorf("failed to initialize sqlite: %w", err)
					}

					// init badger
					var kvStore tokenmanager.ITokenStore

					var badgerDB *badgerdb.DB
					if conf.KvDbEngine == "badgerdb" {
						badgerDbPath := filepath.Join(conf.HubDataDir(), "badger")
						badgerDB, err = badgerdb.New(l, badgerDbPath)
						if err != nil {
							return fmt.Errorf("failed to initialize badgerdb: %w", err)
						}
						kvStore = badgerDB
					} else {
						kvStore = tokenmanager.NewMemoryTokenStore()
					}

					l.Infof("using kv store: %s", conf.KvDbEngine)

					// Print SQLite version if using SQLite storage
					sqliteVersion, err := sqliteDB.GetSQLiteVersion(ctx)
					if err != nil {
						return fmt.Errorf("failed to get SQLite version: %w", err)
					}
					l.Infof("SQLite version: %s", sqliteVersion)

					// check and run all migrations
					m := migrator.New(l, sql.MigrationsFS, sql.MigrationsPath, datamigrations.Migrations)
					if err := m.MigrateUpAll(ctx, sqliteDbPath); err != nil {
						return fmt.Errorf("failed to run migrations: %w", err)
					}

					st, err := store.New(sqliteDB.DB, kvStore)
					if err != nil {
						return fmt.Errorf("failed to initialize store: %w", err)
					}

					systemInfo := models.GetSystemInfo(version, commitHash, buildDate)
					systemInfo.SqliteVersion = sqliteVersion

					// Init receiver
					rc := receiver.New()

					// Initialize dispatcher
					dispatcher := dispatcher.New()

					baseServices := baseservices.New(l, st, authConfig, rc, dispatcher, systemInfo)

					// init alert resolver
					ar := alertresolver.New(l, baseServices)

					// Initialize scheduler
					sched := scheduler.New(l, rc, baseServices, ar)

					availableUpdateData := locker.New(models.AvailableUpdate{})

					// Initialize upgrader if configured
					updater, err := updater.New(l, conf.Updater, version, availableUpdateData)
					if err != nil {
						return fmt.Errorf("failed to initialize upgrader: %w", err)
					}

					srv := servers.New(ctx, l, conf.Server.Addr, baseServices, ar, systemInfo, availableUpdateData)

					// register services
					ln.ServicesRunner().Register(
						service.New(service.WithService(pingpong.New(l))),
						service.New(service.WithService(sqliteDB)),
						service.New(service.WithService(updater)),
						service.New(service.WithService(rc)),
						service.New(service.WithService(dispatcher)),
						service.New(service.WithService(sched)),
						service.New(service.WithService(srv)),
						service.New(service.WithService(baseServices.Notifications().Sender())),
					)

					if badgerDB != nil {
						ln.ServicesRunner().Register(service.New(service.WithService(badgerDB)))
					}

					return ln.Run()
				},
			},
		},
	}
}
