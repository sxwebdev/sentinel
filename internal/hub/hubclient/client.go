package hubclient

import (
	"context"

	"connectrpc.com/connect"
	agentv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agent/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agent/v1/agentv1connect"
	"github.com/sxwebdev/sentinel/internal/hub/hubutils"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/tkcrm/mx/clients/connectrpc_client"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Client struct {
	fingerprint string
	systemInfo  models.SystemInfo

	agentsService agentv1connect.AgentServiceClient
}

func New(
	ctx context.Context,
	l logger.Logger,
	token string,
	fingerprint string,
	systemInfo models.SystemInfo,
	connectRPCConfig connectrpc_client.Config,
) (*Client, error) {
	c := &Client{
		fingerprint: fingerprint,
		systemInfo:  systemInfo,
	}

	agentsService, err := connectrpc_client.New(
		connectRPCConfig, l,
		agentv1connect.NewAgentServiceClient,
		connectrpc_client.WithContext(ctx),
		connectrpc_client.WithConnectrpcOpts(
			connect.WithInterceptors(
				newInterceptor(token),
			),
		),
	)
	if err != nil {
		return nil, err
	}
	c.agentsService = agentsService

	return c, nil
}

// Ping checks the connectivity with the hub server.
func (c *Client) Ping(ctx context.Context) error {
	_, err := c.agentsService.Ping(ctx, connect.NewRequest(&emptypb.Empty{}))
	return err
}

// Authenticate authenticates the client with the hub server.
func (c *Client) Authenticate(ctx context.Context) error {
	_, err := c.agentsService.Authenticate(ctx, connect.NewRequest(&agentv1.AuthenticateRequest{
		Fingerprint: c.fingerprint,
	}))
	return err
}

// ReportSystemInfo reports the system information to the hub server.
func (c *Client) ReportSystemInfo(ctx context.Context) error {
	_, err := c.agentsService.ReportSystemInfo(ctx, connect.NewRequest(&agentv1.ReportSystemInfoRequest{
		SystemInfo: hubutils.ConvertSystemInfoToProto(c.systemInfo),
	}))
	return err
}

// FetchServices fetches the list of services from the hub server.
func (c *Client) FetchServices(ctx context.Context) ([]*models.Service, error) {
	resp, err := c.agentsService.FetchServices(ctx, connect.NewRequest(&agentv1.FetchServicesRequest{}))
	if err != nil {
		return nil, err
	}

	services := make([]*models.Service, 0, len(resp.Msg.Services))
	for _, svcProto := range resp.Msg.Services {
		svc := hubutils.ConvertServiceFromProto(svcProto)
		services = append(services, svc)
	}

	return services, nil
}
