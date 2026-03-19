package apiserver

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	projectsv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/projects/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/projects/v1/projectsv1connect"
	"github.com/sxwebdev/sentinel/internal/servers/ptconverts"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/services/projects"
)

type ProjectsServer struct {
	bs *baseservices.BaseServices

	projectsv1connect.UnimplementedProjectsServiceHandler
}

func newProjectsServer(
	bs *baseservices.BaseServices,
) *ProjectsServer {
	return &ProjectsServer{
		bs: bs,
	}
}

// Name returns the name of the projects server
func (s *ProjectsServer) Name() string { return projectsv1connect.ProjectsServiceName }

// RegisterHandler registers the projects server handler
func (s *ProjectsServer) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return projectsv1connect.NewProjectsServiceHandler(s, opts...)
}

// ProjectsList returns the list of projects
func (s *ProjectsServer) ProjectsList(
	ctx context.Context,
	_ *connect.Request[projectsv1.ProjectsListRequest],
) (*connect.Response[projectsv1.ProjectsListResponse], error) {
	data, err := s.bs.Projects().GetAll(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	projects := make([]*projectsv1.Project, 0, len(data))
	for _, project := range data {
		projects = append(projects, ptconverts.ConvertProjectToProto(project))
	}

	res := connect.NewResponse(&projectsv1.ProjectsListResponse{
		Items: projects,
	})

	return res, nil
}

// ProjectCreate creates a new project
func (s *ProjectsServer) ProjectCreate(
	ctx context.Context,
	req *connect.Request[projectsv1.ProjectCreateRequest],
) (*connect.Response[projectsv1.ProjectCreateResponse], error) {
	params := projects.CreateParams{
		Name:        req.Msg.GetName(),
		Description: req.Msg.GetDescription(),
		Settings:    ptconverts.ConvertProjectSettingsFromProto(req.Msg.GetSettings()),
	}

	data, err := s.bs.Projects().Create(ctx, params)
	if err != nil {
		return nil, newConnectError(err)
	}

	res := connect.NewResponse(&projectsv1.ProjectCreateResponse{
		Item: ptconverts.ConvertProjectToProto(data),
	})

	return res, nil
}

// ProjectGet returns a project by ID
func (s *ProjectsServer) ProjectGet(
	ctx context.Context,
	req *connect.Request[projectsv1.ProjectGetRequest],
) (*connect.Response[projectsv1.ProjectGetResponse], error) {
	data, err := s.bs.Projects().GetByID(ctx, req.Msg.GetId())
	if err != nil {
		return nil, newConnectError(err)
	}
	if data == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	res := connect.NewResponse(&projectsv1.ProjectGetResponse{
		Item: ptconverts.ConvertProjectToProto(data),
	})

	return res, nil
}

// ProjectUpdate updates a project by ID
func (s *ProjectsServer) ProjectUpdate(
	ctx context.Context,
	req *connect.Request[projectsv1.ProjectUpdateRequest],
) (*connect.Response[projectsv1.ProjectUpdateResponse], error) {
	params := projects.UpdateParams{
		ID:          req.Msg.GetId(),
		Name:        req.Msg.GetName(),
		Description: req.Msg.GetDescription(),
		Settings:    ptconverts.ConvertProjectSettingsFromProto(req.Msg.GetSettings()),
	}

	data, err := s.bs.Projects().Update(ctx, params)
	if err != nil {
		return nil, newConnectError(err)
	}
	if data == nil {
		return nil, connect.NewError(connect.CodeNotFound, nil)
	}

	res := connect.NewResponse(&projectsv1.ProjectUpdateResponse{
		Item: ptconverts.ConvertProjectToProto(data),
	})

	return res, nil
}

// ProjectDelete deletes a project by ID
func (s *ProjectsServer) ProjectDelete(
	ctx context.Context,
	req *connect.Request[projectsv1.ProjectDeleteRequest],
) (*connect.Response[projectsv1.ProjectDeleteResponse], error) {
	err := s.bs.Projects().Delete(ctx, req.Msg.GetId())
	if err != nil {
		return nil, newConnectError(err)
	}

	res := connect.NewResponse(&projectsv1.ProjectDeleteResponse{})

	return res, nil
}
