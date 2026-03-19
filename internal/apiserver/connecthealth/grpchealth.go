package connecthealth

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
)

type HealthCheckService struct {
	checker grpchealth.Checker
}

func NewHealthCheckService(serverNames []string) *HealthCheckService {
	checker := grpchealth.NewStaticChecker(serverNames...)
	return &HealthCheckService{
		checker: checker,
	}
}

func (h *HealthCheckService) Name() string {
	return grpchealth.HealthV1ServiceName
}

func (h *HealthCheckService) RegisterHandler(opts ...connect.HandlerOption) (string, http.Handler) {
	return grpchealth.NewHandler(h.checker, opts...)
}

func (h *HealthCheckService) Check(ctx context.Context, req *grpchealth.CheckRequest) (*grpchealth.CheckResponse, error) {
	return h.checker.Check(ctx, req)
}
