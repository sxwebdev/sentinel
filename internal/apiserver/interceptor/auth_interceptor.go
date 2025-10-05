package interceptor

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/sxwebdev/sentinel/internal/hub/hubserver/api/sentinel/auth/v1/authv1connect"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/tkcrm/mx/logger"
)

var availableMethodsBeforeAuth = map[string]struct{}{
	// Auth
	authv1connect.AuthServiceAuthorizationProcedure: {},
	authv1connect.AuthServiceRefreshTokenProcedure:  {},
}

type AuthInterceptor struct {
	logger logger.Logger

	restrictedServices map[string][]userRole

	bs *baseservices.BaseServices
}

func New(
	l logger.Logger,
	bs *baseservices.BaseServices,
) *AuthInterceptor {
	return &AuthInterceptor{
		logger: l,
		bs:     bs,
	}
}

type (
	UserKey             int
	UserOrganizationKey int
)

var (
	UserCtxKey             UserKey
	UserOrganizationCtxKey UserOrganizationKey
)

func (s *AuthInterceptor) ConnectRPCAuthMiddleware() *authn.Middleware {
	authFn := func(ctx context.Context, req *http.Request) (any, error) {
		ud, err := s.authorize(ctx, req)
		if err != nil {
			return nil, err
		}

		return ud, nil
	}

	return authn.NewMiddleware(authFn)
}

func (s *AuthInterceptor) authorize(ctx context.Context, req *http.Request) (*UserDataContext, error) {
	method, ok := authn.InferProcedure(req.URL)
	if !ok {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to infer procedure"))
	}

	// available reflect and health methods
	if strings.Contains(method, grpcreflect.ReflectV1AlphaServiceName) ||
		strings.Contains(method, grpcreflect.ReflectV1ServiceName) ||
		strings.Contains(method, grpchealth.HealthV1ServiceName) {
		return nil, nil //nolint:nilnil
	}

	if _, ok := availableMethodsBeforeAuth[method]; ok {
		return nil, nil //nolint:nilnil
	}

	// get access token from header
	parts := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(parts) < 2 || parts[0] != "Bearer" {
		err := authn.Errorf("expected Bearer authentication scheme")
		return nil, err
	}

	accessToken := parts[1]

	// check access token in auth service
	claims, err := s.bs.Auth().Manager().Authenticate(ctx, accessToken)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid access token: %w", err))
	}

	if claims == nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("empty claims"))
	}

	userID := claims.UserID

	// get user
	user, err := s.bs.Users().GetByID(ctx, userID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("get user id: %w", err))
	}

	ud := &UserDataContext{
		User:        user,
		AccessToken: accessToken,
		Claims:      claims,
	}

	return ud, nil
}
