package apiserver

import (
	"context"
	"fmt"
	"net/http"

	"connectrpc.com/connect"
	authpbv1 "github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/auth/v1"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/auth/v1/authv1connect"
	"github.com/sxwebdev/sentinel/internal/servers/hutils"
	"github.com/sxwebdev/sentinel/internal/servers/ptconverts"
	"github.com/sxwebdev/sentinel/internal/services/auth"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServer struct {
	baseServices *baseservices.BaseServices

	authv1connect.UnimplementedAuthServiceHandler
}

func newAuthServer(
	baseServices *baseservices.BaseServices,
) *AuthServer {
	return &AuthServer{
		baseServices: baseServices,
	}
}

func (s *AuthServer) Name() string { return authv1connect.AuthServiceName }

func (s *AuthServer) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return authv1connect.NewAuthServiceHandler(s, opts...)
}

/*

	Authorization

*/

func (s *AuthServer) Authorization(
	ctx context.Context,
	req *connect.Request[authpbv1.AuthorizationRequest],
) (*connect.Response[authpbv1.AuthorizationResponse], error) {
	req.Msg.Email = hutils.NormalizeEmail(req.Msg.GetEmail())

	if req.Msg.GetEmail() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("email is required"))
	}

	deviceInfo := req.Msg.GetDeviceInfo()
	if deviceInfo == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("device info is required"))
	}

	sessData := auth.SessionData{
		DeviceInfo: auth.DeviceInfo{
			DeviceID:    deviceInfo.GetDeviceId(),
			DeviceType:  convertPbToDeviceType(deviceInfo.GetDeviceType()),
			DeviceName:  deviceInfo.GetDeviceName(),
			Fingerprint: deviceInfo.GetFingerprint(),
			OSVersion:   deviceInfo.GetOsVersion(),
			AppVersion:  deviceInfo.GetAppVersion(),
		},
	}

	data, err := s.baseServices.Auth().Authorization(ctx, req.Msg.GetEmail(), req.Msg.GetPassword(), sessData)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	resp := &authpbv1.AuthorizationResponse{
		AuthPayload: &authpbv1.AuthPayload{
			AccessToken:           data.SessionPair.AccessToken,
			RefreshToken:          data.SessionPair.RefreshToken,
			AccessTokenExpiredAt:  timestamppb.New(data.AccessTokenData.Expiry),
			RefreshTokenExpiredAt: timestamppb.New(data.RefreshTokenData.Expiry),
			DeviceId:              data.AccessTokenData.AdditionalData.DeviceInfo.DeviceID,
		},
		User: ptconverts.ConvertUserToProto(data.User),
	}

	res := connect.NewResponse(resp)

	return res, nil
}

/*
Authenticate
*/
func (s *AuthServer) Authenticate(
	ctx context.Context,
	req *connect.Request[authpbv1.AuthenticateRequest],
) (*connect.Response[authpbv1.AuthenticateResponse], error) {
	if req.Msg.GetAccessToken() == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("access token is required"))
	}

	claims, err := s.baseServices.Auth().Manager().Authenticate(ctx, req.Msg.GetAccessToken())
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid access token: %w", err))
	}

	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("empty claims"))
	}

	// get user
	user, err := s.baseServices.Users().GetByID(ctx, claims.UserID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("get user id: %w", err))
	}

	resp := &authpbv1.AuthenticateResponse{
		User: ptconverts.ConvertUserToProto(user),
	}

	res := connect.NewResponse(resp)

	return res, nil
}

/*

	RefreshToken

*/

func (s *AuthServer) RefreshToken(
	ctx context.Context,
	req *connect.Request[authpbv1.RefreshTokenRequest],
) (*connect.Response[authpbv1.RefreshTokenResponse], error) {
	data, err := s.baseServices.Auth().RefreshToken(ctx, req.Msg.RefreshToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	res := connect.NewResponse(&authpbv1.RefreshTokenResponse{
		AccessToken:           data.SessionPair.AccessToken,
		RefreshToken:          data.SessionPair.RefreshToken,
		AccessTokenExpiredAt:  timestamppb.New(data.AccessTokenData.Expiry),
		RefreshTokenExpiredAt: timestamppb.New(data.RefreshTokenData.Expiry),
	})

	return res, nil
}

// ActiveSessions
func (s *AuthServer) ActiveSessions(
	ctx context.Context,
	_ *connect.Request[authpbv1.ActiveSessionsRequest],
) (*connect.Response[authpbv1.ActiveSessionsResponse], error) {
	userData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	data, err := s.baseServices.Auth().Manager().GetAllSessions(ctx, userData.User.ID)
	if err != nil {
		return nil, newConnectError(err)
	}

	sessions := make([]*authpbv1.Session, 0, len(data))
	for _, session := range data {
		if session.AdditionalData.Country == "" {
			session.AdditionalData.Country = "Unknown"
		}

		if session.AdditionalData.IP == "" {
			session.AdditionalData.IP = "Unknown"
		}

		sessions = append(sessions, &authpbv1.Session{
			DeviceInfo: &authpbv1.DeviceInfo{
				DeviceId:    session.AdditionalData.DeviceInfo.DeviceID,
				DeviceType:  convertDeviceTypeToPb(session.AdditionalData.DeviceInfo.DeviceType),
				DeviceName:  session.AdditionalData.DeviceInfo.DeviceName,
				Fingerprint: session.AdditionalData.DeviceInfo.Fingerprint,
				OsVersion:   session.AdditionalData.DeviceInfo.OSVersion,
				AppVersion:  session.AdditionalData.DeviceInfo.AppVersion,
			},
			Country:   session.AdditionalData.Country,
			Ip:        session.AdditionalData.IP,
			CreatedAt: timestamppb.New(session.AdditionalData.CreatedAt),
		})
	}

	res := connect.NewResponse(&authpbv1.ActiveSessionsResponse{
		Sessions: sessions,
	})

	return res, nil
}

// DeleteSession
func (s *AuthServer) DeleteSession(
	ctx context.Context,
	req *connect.Request[authpbv1.DeleteSessionRequest],
) (*connect.Response[authpbv1.DeleteSessionResponse], error) {
	userData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	if err := s.baseServices.Auth().Manager().DeleteSession(ctx, userData.User.ID, req.Msg.GetDeviceId()); err != nil {
		return nil, newConnectError(err)
	}

	return connect.NewResponse(new(authpbv1.DeleteSessionResponse)), nil
}

// TerminateAllSessions
func (s *AuthServer) TerminateAllSessions(
	ctx context.Context,
	req *connect.Request[authpbv1.TerminateAllSessionsRequest],
) (*connect.Response[authpbv1.TerminateAllSessionsResponse], error) {
	userData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	if err := s.baseServices.Auth().Manager().TerminateAllSessions(ctx, userData.User.ID, req.Msg.GetDeviceId()); err != nil {
		return nil, newConnectError(err)
	}

	return connect.NewResponse(new(authpbv1.TerminateAllSessionsResponse)), nil
}

// Logout
func (s *AuthServer) Logout(
	ctx context.Context,
	_ *connect.Request[authpbv1.LogoutRequest],
) (*connect.Response[authpbv1.LogoutResponse], error) {
	userData, err := getUserDataContext(ctx)
	if err != nil {
		return nil, newConnectError(err)
	}

	if err := s.baseServices.Auth().Manager().Logout(ctx, userData.AccessToken); err != nil {
		return nil, newConnectError(err)
	}

	return connect.NewResponse(new(authpbv1.LogoutResponse)), nil
}

func convertPbToDeviceType(t authpbv1.DeviceType) auth.DeviceType {
	switch t {
	case authpbv1.DeviceType_DEVICE_TYPE_WEB:
		return auth.DeviceTypeWeb
	default:
		return auth.DeviceTypeUnknown
	}
}

func convertDeviceTypeToPb(t auth.DeviceType) authpbv1.DeviceType {
	switch t {
	case auth.DeviceTypeWeb:
		return authpbv1.DeviceType_DEVICE_TYPE_WEB
	default:
		return authpbv1.DeviceType_DEVICE_TYPE_UNSPECIFIED
	}
}
