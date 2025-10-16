package apiserver

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	agentsv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agents/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/agents/v1/agentsv1connect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/servers/ptconverts"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/tkcrm/modules/pkg/db/dbutils"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AgentsServer struct {
	gCtx   context.Context
	logger logger.Logger

	bs *baseservices.BaseServices

	agentsv1connect.UnimplementedAgentsServiceHandler
}

func newAgentsServer(
	gCtx context.Context,
	logger logger.Logger,
	bs *baseservices.BaseServices,
) *AgentsServer {
	return &AgentsServer{
		gCtx:   gCtx,
		logger: logger,
		bs:     bs,
	}
}

// Name returns the name of the agents server
func (s *AgentsServer) Name() string { return agentsv1connect.AgentsServiceName }

// RegisterHandler registers the agents server handler
func (s *AgentsServer) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return agentsv1connect.NewAgentsServiceHandler(s, opts...)
}

// AgentsList lists all agents
func (s *AgentsServer) AgentsList(
	ctx context.Context,
	_ *connect.Request[agentsv1.AgentsListRequest],
) (*connect.Response[agentsv1.AgentsListResponse], error) {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	data, err := s.bs.Agents().Find(ctx, agents.FindParams{
		ProjectID: ctxData.Project.ID,
	})
	if err != nil {
		return nil, newConnectError(err)
	}

	res := connect.NewResponse(&agentsv1.AgentsListResponse{
		Items: make([]*agentsv1.Agent, 0, len(data.Items)),
		Count: data.Count,
	})

	for _, a := range data.Items {
		agent, err := ptconverts.ConvertAgentToProto(a)
		if err != nil {
			return nil, newConnectError(err)
		}
		res.Msg.Items = append(res.Msg.Items, agent)
	}

	return res, nil
}

// AgentsGet returns agent by ID
func (s *AgentsServer) AgentsGet(
	ctx context.Context,
	req *connect.Request[agentsv1.AgentsGetRequest],
) (*connect.Response[agentsv1.AgentsGetResponse], error) {
	item, err := s.bs.Agents().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, newConnectError(err)
	}

	pbitem, err := ptconverts.ConvertAgentToProto(item)
	if err != nil {
		return nil, newConnectError(err)
	}

	return connect.NewResponse(&agentsv1.AgentsGetResponse{Item: pbitem}), nil
}

// AgentsCreate creates a new agent
func (s *AgentsServer) AgentsCreate(
	ctx context.Context,
	req *connect.Request[agentsv1.AgentsCreateRequest],
) (*connect.Response[agentsv1.AgentsCreateResponse], error) {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	params := agents.CreateParams{
		Name:        req.Msg.GetName(),
		Description: req.Msg.Description,
		Kind:        models.AgentKindTypeExternal,
		Tags:        append([]string(nil), req.Msg.GetTags()...),
		Config:      models.AgentConfig{},
		IsEnabled:   req.Msg.GetIsEnabled(),
		ProjectID:   ctxData.Project.ID,
	}

	created, err := s.bs.Agents().Create(ctx, params)
	if err != nil {
		return nil, newConnectError(err)
	}

	cfg, err := ptconverts.ConvertAgentConfigToProto(created.Config)
	if err != nil {
		return nil, newConnectError(err)
	}

	pbitem := &agentsv1.AgentsCreateResponse{
		Id:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		Token:       created.Token,
		TokenHint:   created.TokenHint,
		Status:      ptconverts.ConvertAgentStatusToProto(created.Status),
		IsEnabled:   created.IsEnabled,
		Tags:        created.Tags,
		Config:      cfg,
		CreatedAt:   timestamppb.New(created.CreatedAt),
	}

	return connect.NewResponse(pbitem), nil
}

// AgentsUpdate updates an agent by ID
func (s *AgentsServer) AgentsUpdate(
	ctx context.Context,
	req *connect.Request[agentsv1.AgentsUpdateRequest],
) (*connect.Response[agentsv1.AgentsUpdateResponse], error) {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	params := agents.UpdateParams{
		Name:        req.Msg.GetName(),
		Description: req.Msg.Description,
		IsEnabled:   req.Msg.GetIsEnabled(),
		Tags:        req.Msg.GetTags(),
		Config:      models.AgentConfig{},
		FieldMask: dbutils.FieldMask[repo_agents.ColumnName]{
			repo_agents.ColumnNameAgentsName,
			repo_agents.ColumnNameAgentsDescription,
			repo_agents.ColumnNameAgentsIsEnabled,
			repo_agents.ColumnNameAgentsTags,
			repo_agents.ColumnNameAgentsConfig,
		},
	}

	item, err := s.bs.Agents().Update(ctx, req.Msg.GetId(), ctxData.Project.ID, params)
	if err != nil {
		return nil, newConnectError(err)
	}

	pbitem, err := ptconverts.ConvertAgentToProto(item)
	if err != nil {
		return nil, newConnectError(err)
	}

	return connect.NewResponse(&agentsv1.AgentsUpdateResponse{Item: pbitem}), nil
}

// AgentsDelete deletes an agent by ID
func (s *AgentsServer) AgentsDelete(
	ctx context.Context,
	req *connect.Request[agentsv1.AgentsDeleteRequest],
) (*connect.Response[agentsv1.AgentsDeleteResponse], error) {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	if err := s.bs.Agents().Delete(ctx, req.Msg.GetId(), ctxData.Project.ID); err != nil {
		return nil, newConnectError(err)
	}
	return connect.NewResponse(&agentsv1.AgentsDeleteResponse{}), nil
}

func (s *AgentsServer) AgentsSubscribe(
	ctx context.Context,
	req *connect.Request[agentsv1.AgentsSubscribeRequest],
	stream *connect.ServerStream[agentsv1.AgentsSubscribeResponse],
) error {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return newConnectError(err)
	}

	broker := s.bs.Dispatcher().Agents()
	sub := broker.Subscribe()
	defer broker.Unsubscribe(sub)

	for ctx.Err() == nil {
		select {
		case msg := <-sub:
			if msg.ProjectID != ctxData.Project.ID {
				continue
			}

			if err := stream.Send(&agentsv1.AgentsSubscribeResponse{}); err != nil {
				s.logger.Errorf("failed to send agent subscription update: %v", err)
			}
		case <-ctx.Done():
			return nil
		case <-s.gCtx.Done():
			return nil
		}
	}

	return nil
}
