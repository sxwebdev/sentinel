package main

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/goccy/go-yaml"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/xconfig"
	"github.com/urfave/cli/v3"
)

func cfgPathsFlag() *cli.StringSliceFlag {
	return &cli.StringSliceFlag{
		Name:    "config",
		Aliases: []string{"c"},
		Usage:   "allows you to use your own paths to configuration files",
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
					data := []struct {
						fileName  string
						envPrefix string
						conf      interface{}
					}{
						{
							fileName:  "config.template.yaml",
							envPrefix: envHubPrefix,
							conf:      new(config.ConfigHub),
						},
						{
							fileName:  "config-agent.template.yaml",
							envPrefix: envAgentPrefix,
							conf:      new(config.ConfigAgent),
						},
					}

					for _, d := range data {
						_, err := xconfig.Load(d.conf, xconfig.WithEnvPrefix(d.envPrefix))
						if err != nil {
							return fmt.Errorf("failed to generate markdown: %w", err)
						}

						buf := bytes.NewBuffer(nil)
						enc := yaml.NewEncoder(buf, yaml.Indent(2))
						defer enc.Close()

						if err := enc.Encode(d.conf); err != nil {
							return fmt.Errorf("failed to encode yaml: %w", err)
						}

						if err := os.WriteFile(d.fileName, buf.Bytes(), 0o600); err != nil {
							return fmt.Errorf("failed to write file: %w", err)
						}
					}

					return nil
				},
			},
		},
	}
}
