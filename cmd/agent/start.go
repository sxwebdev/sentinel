package main

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	"github.com/sxwebdev/sentinel/internal/agent"
	"github.com/sxwebdev/sentinel/internal/agent/agentserver"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/handlerutils"
	"github.com/sxwebdev/sentinel/internal/models"
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

func startCMD() *cli.Command {
	return &cli.Command{
		Name:  "start",
		Usage: "start the server",
		Flags: []cli.Flag{cfgPathsFlag()},
		Action: func(ctx context.Context, cl *cli.Command) error {
			conf := new(config.ConfigAgent)
			if err := config.Load(conf, envPrefix, cl.StringSlice("config")); err != nil {
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

			serverInfo := models.GetSystemInfo(version, commitHash, buildDate)

			// init agent service
			ag := agent.New(l, conf, serverInfo)

			// init agent rpc server
			agentServer := agentserver.New(serverInfo)

			rpcServer := connectrpc_transport.NewServer(
				connectrpc_transport.WithLogger(l),
				connectrpc_transport.WithConfig(conf.Server),
				connectrpc_transport.WithServices(agentServer),
				connectrpc_transport.WithServerHandlerWrapper(
					func(h http.Handler) http.Handler {
						return h2c.NewHandler(
							handlerutils.WithCORS(h),
							&http2.Server{})
					},
				),
				connectrpc_transport.WithReflection(
					agentServer.Name(),
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
				service.New(service.WithService(ag)),
				service.New(service.WithService(rpcServer)),
			)

			return ln.Run()
		},
	}
}
