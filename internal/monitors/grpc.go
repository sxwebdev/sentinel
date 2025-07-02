package monitors

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
)

// GRPCMonitor monitors gRPC services
type GRPCMonitor struct {
	BaseMonitor
	serviceName string
	useTLS      bool
	conn        *grpc.ClientConn
}

// NewGRPCMonitor creates a new gRPC monitor
func NewGRPCMonitor(cfg config.ServiceConfig) (*GRPCMonitor, error) {
	monitor := &GRPCMonitor{
		BaseMonitor: NewBaseMonitor(cfg),
		serviceName: getConfigString(cfg.Config, "service_name", ""),
		useTLS:      getConfigBool(cfg.Config, "tls", false),
	}

	// Create gRPC connection
	var opts []grpc.DialOption

	if monitor.useTLS {
		// Use TLS credentials
		tlsConfig := &tls.Config{
			ServerName: getConfigString(cfg.Config, "server_name", ""),
		}
		if getConfigBool(cfg.Config, "insecure_tls", false) {
			tlsConfig.InsecureSkipVerify = true
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		// Use insecure connection
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Add timeout
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(cfg.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	monitor.conn = conn
	return monitor, nil
}

// Check performs the gRPC health check
func (g *GRPCMonitor) Check(ctx context.Context) error {
	// Check connection state
	state := g.conn.GetState()
	if state != connectivity.Ready && state != connectivity.Idle {
		return fmt.Errorf("gRPC connection not ready, state: %s", state)
	}

	// Create health check client
	client := grpc_health_v1.NewHealthClient(g.conn)

	// Prepare health check request
	req := &grpc_health_v1.HealthCheckRequest{}
	if g.serviceName != "" {
		req.Service = g.serviceName
	}

	// Perform health check with timeout
	resp, err := client.Check(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	// Check response status
	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving, status: %s", resp.Status)
	}

	return nil
}

// Close closes the gRPC connection
func (g *GRPCMonitor) Close() error {
	if g.conn != nil {
		return g.conn.Close()
	}
	return nil
}
