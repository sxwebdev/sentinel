package hubserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"github.com/puzpuzpuz/xsync/v3"
	"github.com/sxwebdev/sentinel/internal/alertresolver"
	hubv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/hub/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/hub/v1/hubv1connect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/servers/ptconverts"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/transport/connectrpc_transport"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	logger logger.Logger

	// global context
	gCtx context.Context

	baseservices *baseservices.BaseServices
	ar           *alertresolver.AlertResolver

	// Map of authorized agents: key is agent ID, value is name
	authorizedAgents *xsync.MapOf[string, string]

	hubv1connect.UnimplementedHubServiceHandler
	connectrpc_transport.ConnectRPCService
}

func New(
	gCtx context.Context,
	l logger.Logger,
	baseservices *baseservices.BaseServices,
	ar *alertresolver.AlertResolver,
) *Server {
	return &Server{
		logger:           l,
		gCtx:             gCtx,
		baseservices:     baseservices,
		ar:               ar,
		authorizedAgents: xsync.NewMapOf[string, string](),
	}
}

func (s *Server) Name() string { return hubv1connect.HubServiceName }

func (s *Server) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return hubv1connect.NewHubServiceHandler(s, opts...)
}

// Ping is a health check endpoint.
func (s *Server) Ping(_ context.Context, _ *connect.Request[emptypb.Empty]) (*connect.Response[emptypb.Empty], error) {
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// Authenticate is a no-op for now
func (s *Server) Authenticate(ctx context.Context, req *connect.Request[hubv1.AuthenticateRequest]) (*connect.Response[hubv1.AuthenticateResponse], error) {
	agentData, err := agentDataFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := s.baseservices.Agents().CheckAndUpsertFingerprint(ctx, agentData.Agent.ID, req.Msg.Fingerprint); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	return connect.NewResponse(&hubv1.AuthenticateResponse{}), nil
}

// ReportSystemInfo returns information about the agent
func (s *Server) ReportSystemInfo(ctx context.Context, req *connect.Request[hubv1.ReportSystemInfoRequest]) (*connect.Response[hubv1.ReportSystemInfoResponse], error) {
	agentData, err := agentDataFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if _, err := s.baseservices.Agents().Update(ctx,
		agentData.Agent.ID,
		agentData.Agent.ProjectID,
		agents.UpdateParams{
			Status:     ptconverts.ConvertAgentStatusFromProto(req.Msg.Status),
			SystemInfo: ptconverts.ConvertSystemInfoFromProto(req.Msg.SystemInfo),
			LastSeenAt: utils.Pointer(time.Now()),
			FieldMask: dbutils.FieldMask[repo_agents.ColumnName]{
				repo_agents.ColumnNameAgentsStatus,
				repo_agents.ColumnNameAgentsSystemInfo,
				repo_agents.ColumnNameAgentsLastSeenAt,
			},
		}); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(hubv1.ReportSystemInfoResponse)), nil
}

// FetchServices fetches the list of services to be monitored by the agent.
// func (s *Server) FetchServices(ctx context.Context, _ *connect.Request[hubv1.FetchServicesRequest]) (*connect.Response[hubv1.FetchServicesResponse], error) {
// 	agentData, err := agentDataFromContext(ctx)
// 	if err != nil {
// 		return nil, connect.NewError(connect.CodeUnauthenticated, err)
// 	}

// 	services, err := s.baseservices.Services().GetAllEnabledByAgentID(ctx, agentData.Agent.ID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	var respServices []*servicev1.Service
// 	for _, svc := range services {
// 		svcProto, err := ptconverts.ConvertServiceToProto(svc)
// 		if err != nil {
// 			return nil, err
// 		}
// 		respServices = append(respServices, svcProto)
// 	}

// 	return connect.NewResponse(&hubv1.FetchServicesResponse{
// 		Services: respServices,
// 	}), nil
// }

// SubscribeServices is a server-side streaming RPC that streams service updates to the agent.
func (s *Server) SubscribeServices(
	ctx context.Context,
	req *connect.Request[hubv1.SubscribeServicesRequest],
	stream *connect.ServerStream[hubv1.SubscribeServicesResponse],
) error {
	agentData, err := agentDataFromContext(ctx)
	if err != nil {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	s.authorizedAgents.Store(agentData.Agent.ID, agentData.Agent.Name)
	if _, err := s.baseservices.Agents().Update(ctx,
		agentData.Agent.ID,
		agentData.Agent.ProjectID,
		agents.UpdateParams{
			Status: models.AgentStatusTypeActive,
			FieldMask: dbutils.FieldMask[repo_agents.ColumnName]{
				repo_agents.ColumnNameAgentsStatus,
			},
		}); err != nil {
		return connect.NewError(connect.CodeInternal, err)
	}

	s.logger.Infoln("agent connected:", agentData.Agent.ID, agentData.Agent.Name)
	defer func() {
		if _, err := s.baseservices.Agents().Update(context.Background(),
			agentData.Agent.ID,
			agentData.Agent.ProjectID,
			agents.UpdateParams{
				Status: models.AgentStatusTypeInactive,
				FieldMask: dbutils.FieldMask[repo_agents.ColumnName]{
					repo_agents.ColumnNameAgentsStatus,
				},
			}); err != nil {
			s.logger.Errorf("failed to set agent status to inactive: %s", err)
		}

		s.authorizedAgents.Delete(agentData.Agent.ID)

		s.logger.Infoln("agent disconnected:", agentData.Agent.ID, agentData.Agent.Name)
	}()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// // send test upsert
			// if err := stream.Send(&hubv1.SubscribeServicesResponse{
			// 	Update: &hubv1.SubscribeServicesResponse_Upsert{
			// 		Upsert: &hubv1.ServiceUpsert{
			// 			Service: &servicev1.Service{
			// 				Id: "service-id-upsert",
			// 			},
			// 		},
			// 	},
			// }); err != nil {
			// 	return err
			// }

			// // send test delete
			// if err := stream.Send(&hubv1.SubscribeServicesResponse{
			// 	Update: &hubv1.SubscribeServicesResponse_Delete{
			// 		Delete: &hubv1.ServiceDelete{
			// 			ServiceId: "service-id-delete",
			// 		},
			// 	},
			// }); err != nil {
			// 	return err
			// }
		case <-ctx.Done():
			return nil
		case <-s.gCtx.Done():
			return nil
		}
	}
}

// StreamChecks streams check results from the agent to the server.
func (s *Server) StreamChecks(ctx context.Context, stream *connect.BidiStream[hubv1.StreamChecksRequest, hubv1.StreamChecksResponse]) error {
	s.logger.Infoln("agent started streaming check results")
	for ctx.Err() == nil {
		req, err := stream.Receive()
		if err != nil {
			return err
		}

		if len(req.GetResults()) == 0 {
			continue
		}

		s.logger.Infof("received %d check results, seq: %d", len(req.GetResults()), req.GetSeq())

		for _, result := range req.GetResults() {
			var isSuccess bool
			if result.GetResponseErrorMessage() == "" {
				isSuccess = true
			}

			if isSuccess {
				if err := s.ar.RecordSuccess(
					ctx,
					result.GetServiceId(),
					time.Duration(result.GetResponseTime())*time.Millisecond,
				); err != nil {
					return fmt.Errorf("failed to record success: %w", err)
				}
			} else {
				if err := s.ar.RecordFailure(
					ctx,
					result.GetServiceId(),
					errors.New(result.GetResponseErrorMessage()),
					time.Duration(result.GetResponseTime())*time.Millisecond,
				); err != nil {
					return fmt.Errorf("failed to record failure: %w", err)
				}
			}

			// Acknowledge receipt of the result
			if err := stream.Send(&hubv1.StreamChecksResponse{
				UptoSeq: req.GetSeq(),
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
