package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/urfave/cli/v3"
)

func cfgPathsFlag() *cli.StringFlag {
	return &cli.StringFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Value:   "config.yaml",
		Usage:   "allows you to use your own paths to configuration files. by default it uses config.yaml",
	}
}

func configCMD() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "validate, gen envs and flags for config",
		Commands: []*cli.Command{
			{
				Name:  "genenvs",
				Usage: "generate config yaml template",
				Action: func(_ context.Context, _ *cli.Command) error {
					conf := new(config.Config)

					conf, err := config.Load("")
					if err != nil {
						return fmt.Errorf("failed to load config: %w", err)
					}

					buf := bytes.NewBuffer(nil)
					enc := yaml.NewEncoder(buf, yaml.Indent(2))
					defer enc.Close()

					if err := enc.Encode(conf); err != nil {
						return fmt.Errorf("failed to encode yaml: %w", err)
					}

					if err := os.WriteFile("config.template.yaml", buf.Bytes(), 0o600); err != nil {
						return fmt.Errorf("failed to write file: %w", err)
					}

					return nil
				},
			},
		},
	}
}
