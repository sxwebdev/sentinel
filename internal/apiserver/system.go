package apiserver

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	systemv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/system/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/system/v1/systemv1connect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/servers/ptconverts"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/pkg/locker"
)

type SystemServer struct {
	bs *baseservices.BaseServices

	systemInfo          *models.SystemInfo
	availableUpdateData *locker.Locker[models.AvailableUpdate]

	systemv1connect.UnimplementedSystemServiceHandler
}

func newSystemServer(
	bs *baseservices.BaseServices,
	systemInfo *models.SystemInfo,
	availableUpdateData *locker.Locker[models.AvailableUpdate],
) *SystemServer {
	return &SystemServer{
		bs:                  bs,
		systemInfo:          systemInfo,
		availableUpdateData: availableUpdateData,
	}
}

// Name returns the name of the system server
func (s *SystemServer) Name() string { return systemv1connect.SystemServiceName }

// RegisterHandler registers the system server handler
func (s *SystemServer) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return systemv1connect.NewSystemServiceHandler(s, opts...)
}

// CheckIsInitialized checks if the system is initialized
func (s *SystemServer) CheckIsInitialized(
	ctx context.Context,
	_ *connect.Request[systemv1.CheckIsInitializedRequest],
) (*connect.Response[systemv1.CheckIsInitializedResponse], error) {
	isInitialized, err := s.bs.System().CheckIsInitialized(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&systemv1.CheckIsInitializedResponse{
		IsInitialized: isInitialized,
	})

	return res, nil
}

// Initialize initializes the system
func (s *SystemServer) Initialize(
	ctx context.Context,
	req *connect.Request[systemv1.InitializeRequest],
) (*connect.Response[systemv1.InitializeResponse], error) {
	if req.Msg.GetEmail() == "" || req.Msg.GetPassword() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, nil)
	}

	err := s.bs.System().Initialize(ctx, req.Msg.GetEmail(), req.Msg.GetPassword())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	res := connect.NewResponse(&systemv1.InitializeResponse{})

	return res, nil
}

// GetSystemInfo returns system information
func (s *SystemServer) GetSystemInfo(
	_ context.Context,
	_ *connect.Request[systemv1.GetSystemInfoRequest],
) (*connect.Response[systemv1.GetSystemInfoResponse], error) {
	res := connect.NewResponse(&systemv1.GetSystemInfoResponse{
		Info: ptconverts.ConvertSystemInfoToProto(*s.systemInfo),
	})

	return res, nil
}

// CheckForUpdates checks for available updates
func (s *SystemServer) CheckForUpdates(
	_ context.Context,
	_ *connect.Request[systemv1.CheckForUpdatesRequest],
) (*connect.Response[systemv1.CheckForUpdatesResponse], error) {
	availableUpdate := s.availableUpdateData.Get()

	res := connect.NewResponse(&systemv1.CheckForUpdatesResponse{
		Info: &systemv1.AvailableUpdate{
			CurrentVersion: availableUpdate.CurrentVersion,
			IsAvailable:    availableUpdate.IsAvailable,
			Details: &systemv1.AvailableUpdate_Details{
				IsAvailableManual: availableUpdate.Details.IsAvailableManual,
				TagName:           availableUpdate.Details.TagName,
				Url:               availableUpdate.Details.URL,
				Description:       availableUpdate.Details.Description,
			},
		},
	})

	return res, nil
}
