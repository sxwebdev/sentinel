package apiserver

import (
	"context"

	"connectrpc.com/grpchealth"
	"github.com/sxwebdev/sentinel/internal/apiserver/connecthealth"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agents/v1/agentsv1connect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/notifications/v1/notificationsv1connect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/projects/v1/projectsv1connect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/system/v1/systemv1connect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/pkg/locker"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/transport/connectrpc_transport"
)

type Server struct {
	healthChecker interface {
		grpchealth.Checker
		connectrpc_transport.ConnectRPCService
	}

	systemServer interface {
		systemv1connect.SystemServiceHandler
		connectrpc_transport.ConnectRPCService
	}

	projectsServer interface {
		projectsv1connect.ProjectsServiceHandler
		connectrpc_transport.ConnectRPCService
	}

	agentsServer interface {
		agentsv1connect.AgentsServiceHandler
		connectrpc_transport.ConnectRPCService
	}

	notificationsServer interface {
		notificationsv1connect.NotificationServiceHandler
		connectrpc_transport.ConnectRPCService
	}
}

func New(
	gCtx context.Context,
	logger logger.Logger,
	bs *baseservices.BaseServices,
	systemInfo *models.SystemInfo,
	availableUpdateData *locker.Locker[models.AvailableUpdate],
) *Server {
	s := &Server{
		systemServer:        newSystemServer(bs, systemInfo, availableUpdateData),
		projectsServer:      newProjectsServer(bs),
		agentsServer:        newAgentsServer(gCtx, logger, bs),
		notificationsServer: newNotificationsServer(gCtx, logger, bs),
	}

	s.healthChecker = connecthealth.NewHealthCheckService(s.AllServerNames())

	return s
}

func (s *Server) AllServers() []connectrpc_transport.ConnectRPCService {
	return []connectrpc_transport.ConnectRPCService{
		s.healthChecker,
		s.systemServer,
		s.projectsServer,
		s.agentsServer,
		s.notificationsServer,
	}
}

func (s *Server) AllServerNames() []string {
	res := make([]string, 0, len(s.AllServers()))
	for _, srv := range s.AllServers() {
		if srv == nil {
			continue
		}
		res = append(res, srv.Name())
	}
	return res
}
