package apiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"github.com/sxwebdev/rbacconnect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/services/auth"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/sxwebdev/tokenmanager"
)

type UserDataContext struct {
	User        *models.User
	Project     *models.Project
	AccessToken string
	Claims      *tokenmanager.Data[auth.SessionData]
}

type Middlewares struct {
	bs       *baseservices.BaseServices
	rbacProv *rbacconnect.Provider
}

func NewMiddlewares(
	bs *baseservices.BaseServices,
	rbacProv *rbacconnect.Provider,
) *Middlewares {
	return &Middlewares{
		bs:       bs,
		rbacProv: rbacProv,
	}
}

func (s *Middlewares) Auth() *authn.Middleware {
	authFn := func(ctx context.Context, req *http.Request) (any, error) {
		// infer method from request
		proc, ok := authn.InferProcedure(req.URL)
		if !ok || proc == "" {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to infer procedure"))
		}
		svc, pkg := rbacconnect.SplitProc(proc)

		pol := s.rbacProv.Get()
		if pol == nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("rbac policy is nil"))
		}

		// check if anonymous access is allowed
		if pol.Evaluate(proc, svc, pkg, []rbacconnect.Role{UserRoleAnonymous}).Allowed {
			// yes — skip authentication (anonymous)
			return nil, nil
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

		// get user
		user, err := s.bs.Users().GetByID(ctx, claims.UserID)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("get user id: %w", err))
		}

		ud := &UserDataContext{
			User:        user,
			AccessToken: accessToken,
			Claims:      claims,
		}

		if pid := req.Header.Get("X-Project-ID"); pid != "" {
			project, err := s.bs.Projects().GetByID(ctx, pid)
			if err != nil {
				return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("get project id: %w", err))
			}
			ud.Project = project
		}

		return ud, nil
	}

	return authn.NewMiddleware(authFn)
}
