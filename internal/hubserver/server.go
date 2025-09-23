package hubserver

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	agentv1 "github.com/sxwebdev/sentinel/internal/hubserver/api/sentinel/agent/v1"
	"github.com/sxwebdev/sentinel/internal/hubserver/api/sentinel/agent/v1/agentv1connect"
	servicev1 "github.com/sxwebdev/sentinel/internal/hubserver/api/sentinel/service/v1"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
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
func (s *Server) Authenticate(_ context.Context, _ *connect.Request[agentv1.AuthenticateRequest]) (*connect.Response[agentv1.AuthenticateResponse], error) {
	return connect.NewResponse(&agentv1.AuthenticateResponse{}), nil
}

// ReportSystemInfo returns information about the agent
func (s *Server) ReportSystemInfo(_ context.Context, _ *connect.Request[agentv1.ReportSystemInfoRequest]) (*connect.Response[agentv1.ReportSystemInfoResponse], error) {
	// return connect.NewResponse(&agentv1.ReportSystemInfoResponse{
	// 	ServerInfo: &commonv1.ServerInfo{
	// 		Version:    s.info.Version,
	// 		CommitHash: s.info.CommitHash,
	// 		BuildDate:  s.info.BuildDate,
	// 		GoVersion:  s.info.GoVersion,
	// 		Os:         s.info.OS,
	// 		Arch:       s.info.Arch,
	// 		StartedAt:  timestamppb.New(s.startedAt),
	// 	},
	// 	Status: agentv1.AgentStatus_AGENT_STATUS_ACTIVE,
	// }), nil

	return connect.NewResponse(new(agentv1.ReportSystemInfoResponse)), nil
}

// FetchServices fetches the list of services to be monitored by the agent.
func (s *Server) FetchServices(_ context.Context, _ *connect.Request[agentv1.FetchServicesRequest]) (*connect.Response[agentv1.FetchServicesResponse], error) {
	agentID := "todo"
	services, err := s.baseservices.Services().GetAllEnabledByAgentID(context.Background(), agentID)
	if err != nil {
		return nil, err
	}

	var respServices []*servicev1.Service
	for _, svc := range services {
		respServices = append(respServices, convertServiceToProto(svc))
	}

	return connect.NewResponse(&agentv1.FetchServicesResponse{
		Services: respServices,
	}), nil
}
