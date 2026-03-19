package agent

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"connectrpc.com/connect"
	hubv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/hub/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/hub/v1/hubv1connect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/servers/ptconverts"
	"github.com/tkcrm/mx/clients/connectrpc_client"
	"github.com/tkcrm/mx/logger"
	"golang.org/x/net/http2"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type client struct {
	logger logger.Logger

	gCtx context.Context

	fingerprint string
	systemInfo  models.SystemInfo

	changeStateCh  chan<- connectionState
	checkResultsCh <-chan checkResult

	agentsService hubv1connect.HubServiceClient
}

func newClient(
	gCtx context.Context,
	l logger.Logger,
	token string,
	fingerprint string,
	systemInfo models.SystemInfo,
	connectRPCConfig connectrpc_client.Config,
	changeStateCh chan<- connectionState,
	checkResultsCh <-chan checkResult,
) (*client, error) {
	c := &client{
		logger:         l,
		gCtx:           gCtx,
		fingerprint:    fingerprint,
		systemInfo:     systemInfo,
		changeStateCh:  changeStateCh,
		checkResultsCh: checkResultsCh,
	}

	h2cTransport := &http2.Transport{
		AllowHTTP: true,
		DialTLS: func(network, addr string, _ *tls.Config) (net.Conn, error) {
			return net.Dial(network, addr)
		},
	}

	agentsService, err := connectrpc_client.New(
		connectRPCConfig, l,
		hubv1connect.NewHubServiceClient,
		connectrpc_client.WithContext(gCtx),
		connectrpc_client.WithConnectrpcOpts(
			connect.WithInterceptors(
				newInterceptor(token),
			),
		),
		connectrpc_client.WithHttpClient(
			&http.Client{Transport: h2cTransport},
		),
	)
	if err != nil {
		return nil, err
	}
	c.agentsService = agentsService

	return c, nil
}

// Ping checks the connectivity with the hub server.
func (c *client) Ping(ctx context.Context) error {
	_, err := c.agentsService.Ping(ctx, connect.NewRequest(&emptypb.Empty{}))
	return err
}

// Authenticate authenticates the client with the hub server.
func (c *client) Authenticate(ctx context.Context) error {
	_, err := c.agentsService.Authenticate(ctx, connect.NewRequest(&hubv1.AuthenticateRequest{
		Fingerprint: c.fingerprint,
	}))
	return err
}

// ReportSystemInfo reports the system information to the hub server.
func (c *client) ReportSystemInfo(ctx context.Context) error {
	_, err := c.agentsService.ReportSystemInfo(ctx, connect.NewRequest(&hubv1.ReportSystemInfoRequest{
		Status:     ptconverts.ConvertAgentStatusToProto(models.AgentStatusTypeActive),
		SystemInfo: ptconverts.ConvertSystemInfoToProto(c.systemInfo),
	}))
	return err
}

// FetchServices fetches the list of services from the hub server.
// func (c *client) FetchServices(ctx context.Context) ([]*models.Service, error) {
// 	resp, err := c.agentsService.FetchServices(ctx, connect.NewRequest(&hubv1.FetchServicesRequest{}))
// 	if err != nil {
// 		return nil, err
// 	}

// 	services := make([]*models.Service, 0, len(resp.Msg.Services))
// 	for _, svcProto := range resp.Msg.Services {
// 		svc, err := ptconverts.ConvertServiceFromProto(svcProto)
// 		if err != nil {
// 			return nil, err
// 		}

// 		services = append(services, svc)
// 	}

// 	return services, nil
// }

// SubscribeServices subscribes to service updates from the hub server.
// func (c *client) SubscribeServices(ctx context.Context, handler func(eventType subscribeServiceType, svcID string, svc *models.Service)) error {
// 	c.logger.Infoln("subscribed to service updates")
// 	defer func() {
// 		c.changeStateCh <- connectionStateDisconnected
// 		c.logger.Infoln("unsubscribing from service updates")
// 	}()

// 	stream, err := c.agentsService.SubscribeServices(ctx, connect.NewRequest(&hubv1.SubscribeServicesRequest{}))
// 	if err != nil {
// 		return err
// 	}

// 	defer func() {
// 		if err := stream.Close(); err != nil {
// 			c.logger.Errorf("failed to close stream: %s", err)
// 		}
// 	}()

// 	for stream.Receive() {
// 		msg := stream.Msg()

// 		eventType := subscribeServiceTypeUpsert
// 		var svcID string
// 		if msg.GetDelete() != nil {
// 			eventType = subscribeServiceTypeDelete
// 			svcID = msg.GetDelete().GetServiceId()
// 		}

// 		var svc *models.Service
// 		if msg.GetUpsert() != nil {
// 			svcID = msg.GetUpsert().GetService().GetId()
// 			svc, err = ptconverts.ConvertServiceFromProto(msg.GetUpsert().GetService())
// 			if err != nil {
// 				return err
// 			}
// 		}

// 		handler(eventType, svcID, svc)
// 	}

// 	return nil
// }

// StreamChecks streams check results to the hub server.
func (c *client) StreamChecks(ctx context.Context) error {
	stream := c.agentsService.StreamChecks(ctx)

	defer stream.CloseResponse()

	errChan := make(chan error, 1)
	// confirmedChan := make(chan uint64, 1)
	var lastSentSequence atomic.Uint64
	var lastConfirmedSequence atomic.Uint64

	// handle confirmations
	go func() {
		for c.gCtx.Err() == nil {
			res, err := stream.Receive()
			if err != nil {
				errChan <- err
			}

			// confirmedChan <- res.GetUptoSeq()
			lastConfirmedSequence.Store(res.GetUptoSeq())
		}
	}()

	// send check results
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-c.gCtx.Done():
			return nil
		case err := <-errChan:
			// try to send last message again with delay
			return err
			// 	case <-confirmedChan:
			// try to send next chunk with results
		case res := <-c.checkResultsCh:
			for c.gCtx.Err() == nil {
				if lastSentSequence.Load() != lastConfirmedSequence.Load() {
					time.Sleep(5 * time.Millisecond)
					continue
				}

				err := stream.Send(&hubv1.StreamChecksRequest{
					Seq: lastSentSequence.Load() + 1,
					Results: []*hubv1.CheckResult{
						{
							ServiceId:            res.serviceID,
							Status:               ptconverts.ConvertServiceStatusToProto(res.status),
							CheckedAt:            timestamppb.New(res.checkedAt),
							ResponseMessage:      res.responseMessage,
							ResponseErrorMessage: res.responseErrorMessage,
							ResponseTime:         int64(res.responseTime.Milliseconds()),
						},
					},
				})
				if err != nil {
					// try to send again
					c.logger.Errorf("failed to send check result, retrying: %s", err)
					time.Sleep(1 * time.Second)
					continue
				}

				lastSentSequence.Add(1)
				break
			}
		}
	}
}
