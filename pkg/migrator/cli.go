package migrator

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v3"
)

// CliCmd returns a cli.Command for managing database migrations.
func CliCmd(l logger, fs embed.FS, migrationsPath string, opsmigrations DataMigrations) *cli.Command {
	return &cli.Command{
		Name:  "migrations",
		Usage: "manage database migrations",
		Commands: []*cli.Command{
			{
				Name:  "create",
				Usage: "create a new database migration",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Aliases:  []string{"n"},
						Usage:    "name of the migration",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "migrations-path",
						Aliases:  []string{"p"},
						Usage:    "path to the migrations folder",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cl *cli.Command) error {
					name := cl.String("name")
					path := cl.String("migrations-path")

					// check if path exists, if not create it
					if _, err := os.Stat(path); os.IsNotExist(err) {
						if err := os.MkdirAll(path, 0o700); err != nil {
							return fmt.Errorf("failed to create migrations dir: %w", err)
						}
					}

					// get max version from existing migrations
					maxVersion, err := GetMaxVersion(path)
					if err != nil {
						return fmt.Errorf("failed to get max version: %w", err)
					}

					nextVersion := maxVersion + 1

					upFile := filepath.Join(path, fmt.Sprintf("%d_%s.up.sql", nextVersion, name))
					downFile := filepath.Join(path, fmt.Sprintf("%d_%s.down.sql", nextVersion, name))

					upContent := "-- SQL in section 'Up' is executed when this migration is applied.\n\n"
					downContent := "-- SQL in section 'Down' is executed when this migration is rolled back.\n\n"

					if err := os.WriteFile(upFile, []byte(upContent), 0o644); err != nil {
						return fmt.Errorf("failed to create up migration file: %w", err)
					}

					if err := os.WriteFile(downFile, []byte(downContent), 0o644); err != nil {
						return fmt.Errorf("failed to create down migration file: %w", err)
					}

					fmt.Printf("Created migration files:\n%s\n%s\n", upFile, downFile)

					return nil
				},
			},
			{
				Name:  "up",
				Usage: "apply the next database migrations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-path",
						Aliases:  []string{"dp"},
						Usage:    "path to database file",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cl *cli.Command) error {
					m := New(l, fs, migrationsPath, opsmigrations)
					return m.MigrateUpAll(ctx, cl.String("db-path"))
				},
			},
			{
				Name:  "down",
				Usage: "roll back the last database migration",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "db-path",
						Aliases:  []string{"dp"},
						Usage:    "path to database file",
						Required: true,
					},
				},
				Action: func(ctx context.Context, cl *cli.Command) error {
					m := New(l, fs, migrationsPath, opsmigrations)
					return m.MigrateDown(ctx, cl.String("db-path"))
				},
			},
		},
	}
}
