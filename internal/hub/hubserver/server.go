package hubserver

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	agentv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agent/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agent/v1/agentv1connect"
	servicev1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/service/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubutils"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
	"github.com/tkcrm/mx/transport/connectrpc_transport"
)

type Server struct {
	baseservices *baseservices.BaseServices

	agentv1connect.UnimplementedAgentServiceHandler
	connectrpc_transport.ConnectRPCService
}

func New(baseservices *baseservices.BaseServices) *Server {
	return &Server{
		baseservices: baseservices,
	}
}

func (s *Server) Name() string { return agentv1connect.AgentServiceName }

func (s *Server) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return agentv1connect.NewAgentServiceHandler(s, opts...)
}

// Authenticate is a no-op for now
func (s *Server) Authenticate(ctx context.Context, req *connect.Request[agentv1.AuthenticateRequest]) (*connect.Response[agentv1.AuthenticateResponse], error) {
	agentData, err := agentDataFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := s.baseservices.Agents().CheckAndUpsertFingerprint(ctx, agentData.Agent.ID, req.Msg.Fingerprint); err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	return connect.NewResponse(&agentv1.AuthenticateResponse{}), nil
}

// ReportSystemInfo returns information about the agent
func (s *Server) ReportSystemInfo(ctx context.Context, req *connect.Request[agentv1.ReportSystemInfoRequest]) (*connect.Response[agentv1.ReportSystemInfoResponse], error) {
	agentData, err := agentDataFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	systemInfo := models.SystemInfo{
		Version:       req.Msg.SystemInfo.Version,
		CommitHash:    req.Msg.SystemInfo.CommitHash,
		BuildDate:     req.Msg.SystemInfo.BuildDate,
		GoVersion:     req.Msg.SystemInfo.GoVersion,
		OS:            req.Msg.SystemInfo.Os,
		Arch:          req.Msg.SystemInfo.Arch,
		Hostname:      req.Msg.SystemInfo.Hostname,
		KernelVersion: req.Msg.SystemInfo.KernelVersion,
		IpAddress:     req.Msg.SystemInfo.IpAddress,
		CpuModel:      req.Msg.SystemInfo.CpuModel,
	}

	if _, err := s.baseservices.Agents().Update(ctx, agentData.Agent.ID, agents.UpdateParams{
		Status:       hubutils.ConvertAgentStatusFromProto(req.Msg.Status),
		SystemInfo:   systemInfo,
		LastOnlineAt: utils.Pointer(time.Now()),
		FieldMask: dbutils.FieldMask[repo_agents.ColumnName]{
			repo_agents.ColumnNameAgentsStatus,
			repo_agents.ColumnNameAgentsSystemInfo,
			repo_agents.ColumnNameAgentsLastOnlineAt,
		},
	}); err != nil {
		return nil, err
	}

	return connect.NewResponse(new(agentv1.ReportSystemInfoResponse)), nil
}

// FetchServices fetches the list of services to be monitored by the agent.
func (s *Server) FetchServices(ctx context.Context, _ *connect.Request[agentv1.FetchServicesRequest]) (*connect.Response[agentv1.FetchServicesResponse], error) {
	agentData, err := agentDataFromContext(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	services, err := s.baseservices.Services().GetAllEnabledByAgentID(ctx, agentData.Agent.ID)
	if err != nil {
		return nil, err
	}

	var respServices []*servicev1.Service
	for _, svc := range services {
		respServices = append(respServices, hubutils.ConvertServiceToProto(svc))
	}

	return connect.NewResponse(&agentv1.FetchServicesResponse{
		Services: respServices,
	}), nil
}
