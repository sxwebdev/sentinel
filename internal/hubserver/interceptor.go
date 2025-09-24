package hubserver

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"connectrpc.com/authn"
	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/services/baseservices"
	"github.com/tkcrm/mx/logger"
)

type AgentDataContext struct {
	Agent models.Agent
}

type Interceptor struct {
	logger logger.Logger

	baseServices *baseservices.BaseServices
}

func NewInterceptor(
	l logger.Logger,
	baseServices *baseservices.BaseServices,
) *Interceptor {
	return &Interceptor{
		logger:       l,
		baseServices: baseServices,
	}
}

type (
	AgentKey int
)

var AgentCtxKey AgentKey

func (s *Interceptor) ConnectRPCAuthMiddleware() *authn.Middleware {
	authFn := func(ctx context.Context, req *http.Request) (any, error) {
		ud, err := s.authorize(ctx, req)
		if err != nil {
			return nil, err
		}

		return ud, nil
	}

	return authn.NewMiddleware(authFn)
}

func (s *Interceptor) authorize(_ context.Context, req *http.Request) (*AgentDataContext, error) {
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

	// get access token from header
	parts := strings.SplitN(req.Header.Get("Authorization"), " ", 2)
	if len(parts) < 2 || parts[0] != "Sentinel" {
		return nil, authn.Errorf("expected Sentinel authentication scheme")
	}

	token := parts[1]

	agentID, secret, err := agents.ParseAuthHeader(token)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("failed to parse auth header: %w", err))
	}

	agent, err := s.baseServices.Agents().GetByID(req.Context(), agentID)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("failed to get agent by ID: %w", err))
	}

	ok, err = agents.VerifySecret(secret, agent.SecretHash)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("failed to verify agent secret: %w", err))
	}
	if !ok {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("agent secret is invalid"))
	}

	if !agent.IsEnabled {
		return nil, connect.NewError(connect.CodePermissionDenied, fmt.Errorf("agent is disabled"))
	}

	ad := &AgentDataContext{
		Agent: *agent,
	}

	return ad, nil
}

func agentDataFromContext(ctx context.Context) (*AgentDataContext, error) {
	u, ok := authn.GetInfo(ctx).(*AgentDataContext)
	if !ok || u == nil {
		return nil, ErrUnavailableAgent
	}

	return u, nil
}
