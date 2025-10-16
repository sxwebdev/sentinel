package apiserver

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	notificationsv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/notifications/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/notifications/v1/notificationsv1connect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/servers/ptconverts"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/sentinel/internal/services/notifications"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/tkcrm/mx/logger"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type NotificationsServer struct {
	gCtx   context.Context
	logger logger.Logger

	bs *baseservices.BaseServices

	notificationsv1connect.UnimplementedNotificationServiceHandler
}

func newNotificationsServer(
	gCtx context.Context,
	logger logger.Logger,
	bs *baseservices.BaseServices,
) *NotificationsServer {
	return &NotificationsServer{
		gCtx:   gCtx,
		logger: logger,
		bs:     bs,
	}
}

// Name returns the name of the notifications server
func (s *NotificationsServer) Name() string { return notificationsv1connect.NotificationServiceName }

// RegisterHandler registers the notifications server handler
func (s *NotificationsServer) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return notificationsv1connect.NewNotificationServiceHandler(s, opts...)
}

// ProvidersList lists all notification providers
func (s *NotificationsServer) ProvidersList(
	ctx context.Context,
	_ *connect.Request[notificationsv1.ProvidersListRequest],
) (*connect.Response[notificationsv1.ProvidersListResponse], error) {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	data, err := s.bs.Notifications().Providers().GetAll(ctx, ctxData.Project.ID)
	if err != nil {
		return nil, newConnectError(err)
	}

	res := connect.NewResponse(&notificationsv1.ProvidersListResponse{
		Items: make([]*notificationsv1.Provider, 0, len(data)),
	})

	for _, p := range data {
		provider, err := convertNotificationProviderToProto(p)
		if err != nil {
			return nil, newConnectError(err)
		}
		res.Msg.Items = append(res.Msg.Items, provider)
	}

	return res, nil
}

// ProviderCreate creates a new notification provider
func (s *NotificationsServer) ProviderCreate(
	ctx context.Context,
	req *connect.Request[notificationsv1.ProviderCreateRequest],
) (*connect.Response[notificationsv1.ProviderCreateResponse], error) {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	params := buildCreateParamsFromProto(req.Msg)
	params.ProjectID = ctxData.Project.ID

	created, err := s.bs.Notifications().Providers().Create(ctx, params)
	if err != nil {
		return nil, newConnectError(err)
	}

	p, err := convertNotificationProviderToProto(created)
	if err != nil {
		return nil, newConnectError(err)
	}

	res := connect.NewResponse(&notificationsv1.ProviderCreateResponse{Item: p})
	return res, nil
}

// ProviderUpdate updates an existing notification provider
func (s *NotificationsServer) ProviderUpdate(
	ctx context.Context,
	req *connect.Request[notificationsv1.ProviderUpdateRequest],
) (*connect.Response[notificationsv1.ProviderUpdateResponse], error) {
	id := req.Msg.GetId()
	params, err := buildUpdateParamsFromProto(req.Msg)
	if err != nil {
		return nil, newConnectError(err)
	}

	updated, err := s.bs.Notifications().Providers().Update(ctx, id, params)
	if err != nil {
		return nil, newConnectError(err)
	}

	p, err := convertNotificationProviderToProto(updated)
	if err != nil {
		return nil, newConnectError(err)
	}

	res := connect.NewResponse(&notificationsv1.ProviderUpdateResponse{Item: p})
	return res, nil
}

// ProviderDelete deletes a notification provider by ID
func (s *NotificationsServer) ProviderDelete(
	ctx context.Context,
	req *connect.Request[notificationsv1.ProviderDeleteRequest],
) (*connect.Response[notificationsv1.ProviderDeleteResponse], error) {
	id := req.Msg.GetId()
	if err := s.bs.Notifications().Providers().Delete(ctx, id); err != nil {
		return nil, newConnectError(err)
	}
	res := connect.NewResponse(&notificationsv1.ProviderDeleteResponse{})
	return res, nil
}

// ProviderTest sends a test notification using the specified provider ID
func (s *NotificationsServer) ProviderTest(
	ctx context.Context,
	req *connect.Request[notificationsv1.ProviderTestRequest],
) (*connect.Response[notificationsv1.ProviderTestResponse], error) {
	id := req.Msg.GetId()
	if err := s.bs.Notifications().Providers().Test(ctx, id); err != nil {
		return nil, newConnectError(err)
	}
	res := connect.NewResponse(&notificationsv1.ProviderTestResponse{})
	return res, nil
}

// HistoryList retrieves notification history with optional filters
func (s *NotificationsServer) HistoryList(
	ctx context.Context,
	req *connect.Request[notificationsv1.HistoryListRequest],
) (*connect.Response[notificationsv1.HistoryListResponse], error) {
	var page, pageSize *uint32
	var orderBy string
	if c := req.Msg.GetCommon(); c != nil {
		if c.Page != nil {
			page = c.Page
		}
		if c.PageSize != nil {
			pageSize = c.PageSize
		}
		if c.OrderBy != nil {
			orderBy = *c.OrderBy
		}
	}

	data, err := s.bs.Notifications().History().Find(ctx, notifications.FindHistoryParams{
		Status:   req.Msg.GetStatus(),
		OrderBy:  orderBy,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return nil, newConnectError(err)
	}

	res := connect.NewResponse(&notificationsv1.HistoryListResponse{
		Items: make([]*notificationsv1.HistoryItem, 0, len(data.Items)),
		Count: data.Count,
	})

	for _, h := range data.Items {
		res.Msg.Items = append(res.Msg.Items, ptconverts.ConvertHistoryViewToProto(h))
	}
	return res, nil
}

func convertNotificationProviderToProto(p *models.NotificationProvider) (*notificationsv1.Provider, error) {
	config := &notificationsv1.ProviderConfig{}
	switch p.ProviderType {
	case models.NotificationProviderTypeShoutrrr:
		cfg := &models.NotificationProviderShoutrrrConfig{}
		if err := p.Config.ConvertToAny(cfg); err != nil {
			return nil, err
		}
		config.Config = &notificationsv1.ProviderConfig_Shoutrrr{
			Shoutrrr: &notificationsv1.ShoutrrrConfig{
				Url: cfg.URL,
			},
		}
	}

	return &notificationsv1.Provider{
		Id:        p.ID,
		Type:      convertNotificationProviderTypeToProto(p.ProviderType),
		Config:    config,
		IsEnabled: p.IsEnabled,
		CreatedAt: timestamppb.New(p.CreatedAt),
		UpdatedAt: timestamppb.New(p.UpdatedAt),
	}, nil
}

func convertNotificationProviderTypeToProto(t models.NotificationProviderType) notificationsv1.ProviderType {
	switch t {
	case models.NotificationProviderTypeShoutrrr:
		return notificationsv1.ProviderType_PROVIDER_TYPE_SHOUTRRR
	default:
		return notificationsv1.ProviderType_PROVIDER_TYPE_UNSPECIFIED
	}
}

// convertNotificationProviderTypeFromProto converts proto enum to internal model type
func convertNotificationProviderTypeFromProto(t notificationsv1.ProviderType) models.NotificationProviderType {
	switch t {
	case notificationsv1.ProviderType_PROVIDER_TYPE_SHOUTRRR:
		return models.NotificationProviderTypeShoutrrr
	default:
		return ""
	}
}

// buildCreateParamsFromProto builds CreateProviderParams from proto request
func buildCreateParamsFromProto(req *notificationsv1.ProviderCreateRequest) notifications.CreateProviderParams {
	cfg := map[string]any{}
	if c := req.GetConfig(); c != nil {
		switch x := c.Config.(type) {
		case *notificationsv1.ProviderConfig_Shoutrrr:
			if x.Shoutrrr != nil {
				cfg["url"] = x.Shoutrrr.GetUrl()
			}
		}
	}

	return notifications.CreateProviderParams{
		ProviderType: convertNotificationProviderTypeFromProto(req.GetType()),
		Config:       cfg,
		IsEnabled:    req.GetIsEnabled(),
	}
}

// buildUpdateParamsFromProto builds UpdateProviderParams from proto request
func buildUpdateParamsFromProto(req *notificationsv1.ProviderUpdateRequest) (notifications.UpdateProviderParams, error) {
	// Convert config oneof into JSONField
	m := map[string]any{}
	if c := req.GetConfig(); c != nil {
		switch x := c.Config.(type) {
		case *notificationsv1.ProviderConfig_Shoutrrr:
			if x.Shoutrrr != nil {
				m["url"] = x.Shoutrrr.GetUrl()
			}
		}
	}

	jf := storecmn.JSONField("{}")
	if err := jf.UnmarshalFromAny(m); err != nil {
		return notifications.UpdateProviderParams{}, err
	}

	return notifications.UpdateProviderParams{
		ProviderType: convertNotificationProviderTypeFromProto(req.GetType()),
		Config:       jf,
		IsEnabled:    req.GetIsEnabled(),
	}, nil
}

// HistorySubscribe streams real-time notification history updates to the client
func (s *NotificationsServer) HistorySubscribe(
	ctx context.Context,
	req *connect.Request[notificationsv1.HistorySubscribeRequest],
	stream *connect.ServerStream[notificationsv1.HistorySubscribeResponse],
) error {
	ctxData, err := getUserDataContext(ctx)
	if err != nil {
		return newConnectError(err)
	}

	broker := s.bs.Dispatcher().NotificationHistory()
	sub := broker.Subscribe()
	defer broker.Unsubscribe(sub)

	for ctx.Err() == nil {
		select {
		case msg := <-sub:
			if msg.ProjectID != ctxData.Project.ID {
				continue
			}

			if err := stream.Send(&notificationsv1.HistorySubscribeResponse{}); err != nil {
				s.logger.Errorf("failed to send notification history update: %v", err)
			}
		case <-ctx.Done():
			return nil
		case <-s.gCtx.Done():
			return nil
		}
	}

	return nil
}
