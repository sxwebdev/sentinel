package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sxwebdev/sentinel/internal/agent"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/tkcrm/mx/launcher"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/service"
	"github.com/tkcrm/mx/service/pingpong"
	"github.com/urfave/cli/v3"
)

func agentCMD() *cli.Command {
	return &cli.Command{
		Name:  "agent",
		Usage: "agent commands",
		Commands: []*cli.Command{
			{
				Name:  "start",
				Usage: "start the agent",
				Flags: []cli.Flag{cfgPathsFlag()},
				Action: func(ctx context.Context, cl *cli.Command) error {
					conf := new(config.ConfigAgent)
					if err := config.Load(conf, envAgentPrefix, cl.StringSlice("config")); err != nil {
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
					if _, err := os.Stat(conf.AgentDataDir()); os.IsNotExist(err) {
						l.Infof("creating data directory in %s", conf.AgentDataDir())
						if err := os.MkdirAll(conf.AgentDataDir(), 0o700); err != nil {
							return fmt.Errorf("failed to create data dir: %w", err)
						}
					}

					serverInfo := models.GetSystemInfo(version, commitHash, buildDate)

					// init agent service
					ag, err := agent.New(ctx, l, conf, serverInfo)
					if err != nil {
						return fmt.Errorf("failed to init agent: %w", err)
					}

					// register services
					ln.ServicesRunner().Register(
						service.New(service.WithService(pingpong.New(l))),
						service.New(service.WithService(ag)),
					)

					return ln.Run()
				},
			},
		},
	}
}
