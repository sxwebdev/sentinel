package agentserver

import (
	"context"
	"net/http"
	"time"

	"connectrpc.com/connect"
	agentv1 "github.com/sxwebdev/sentinel/internal/agent/api/sentinel/agent/v1"
	"github.com/sxwebdev/sentinel/internal/agent/api/sentinel/agent/v1/agentv1connect"
	commonv1 "github.com/sxwebdev/sentinel/internal/agent/api/sentinel/common/v1"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/tkcrm/mx/transport/connectrpc_transport"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	info      models.SystemInfo
	startedAt time.Time

	agentv1connect.UnimplementedAgentServiceHandler
	connectrpc_transport.ConnectRPCService
}

func New(info models.SystemInfo) *Server {
	return &Server{
		info:      info,
		startedAt: time.Now(),
	}
}

func (s *Server) Name() string { return agentv1connect.AgentServiceName }

func (s *Server) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return agentv1connect.NewAgentServiceHandler(s, opts...)
}

// Info returns information about the agent
func (s *Server) Info(_ context.Context, _ *connect.Request[agentv1.InfoRequest]) (*connect.Response[agentv1.InfoResponse], error) {
	return connect.NewResponse(&agentv1.InfoResponse{
		ServerInfo: &commonv1.ServerInfo{
			Version:    s.info.Version,
			CommitHash: s.info.CommitHash,
			BuildDate:  s.info.BuildDate,
			GoVersion:  s.info.GoVersion,
			Os:         s.info.OS,
			Arch:       s.info.Arch,
			StartedAt:  timestamppb.New(s.startedAt),
		},
		Status: agentv1.AgentStatus_AGENT_STATUS_ACTIVE,
	}), nil
}
