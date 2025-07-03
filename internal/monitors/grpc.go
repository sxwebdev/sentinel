package monitors

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

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
	checkType   string // "health", "reflection", "connectivity"
}

// NewGRPCMonitor creates a new gRPC monitor
func NewGRPCMonitor(cfg config.ServiceConfig) (*GRPCMonitor, error) {
	monitor := &GRPCMonitor{
		BaseMonitor: NewBaseMonitor(cfg),
		serviceName: getConfigString(cfg.Config, "service_name", ""),
		useTLS:      getConfigBool(cfg.Config, "tls", false),
		checkType:   getConfigString(cfg.Config, "check_type", "health"),
	}

	// Create gRPC connection options
	var opts []grpc.DialOption

	if monitor.useTLS {
		// Use TLS credentials
		tlsConfig := &tls.Config{}
		if getConfigBool(cfg.Config, "insecure_tls", false) {
			tlsConfig.InsecureSkipVerify = true
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	} else {
		// Use insecure connection
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// Create connection without blocking
	conn, err := grpc.NewClient(cfg.Endpoint, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection: %w", err)
	}

	monitor.conn = conn
	return monitor, nil
}

// Check performs the gRPC health check
func (g *GRPCMonitor) Check(ctx context.Context) error {
	// Check connection state first
	if err := g.checkConnectionState(ctx); err != nil {
		return err
	}

	// Perform check based on type
	switch g.checkType {
	case "health":
		return g.performHealthCheck(ctx)
	case "reflection":
		return g.performReflectionCheck(ctx)
	case "connectivity":
		return nil
	default:
		return fmt.Errorf("unsupported check type: %s", g.checkType)
	}
}

// checkConnectionState verifies the gRPC connection is ready
func (g *GRPCMonitor) checkConnectionState(ctx context.Context) error {
	state := g.conn.GetState()

	// If connecting, wait for state change with timeout
	if state == connectivity.Connecting {
		connectCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if !g.conn.WaitForStateChange(connectCtx, state) {
			return fmt.Errorf("gRPC connection timeout, state: %s", state)
		}
		state = g.conn.GetState()
	}

	// Only Ready state is considered successful
	// Idle state might indicate connection issues for non-existent servers
	if state == connectivity.Ready {
		return nil
	}

	// For other states, try to force a connection attempt and wait
	if state == connectivity.Idle {
		// Force connection attempt
		g.conn.Connect()

		// Wait for state change with timeout
		connectCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		if g.conn.WaitForStateChange(connectCtx, state) {
			state = g.conn.GetState()
			if state == connectivity.Ready {
				return nil
			}
		}

		return fmt.Errorf("gRPC connection failed to establish, final state: %s", state)
	}

	// Handle other states
	switch state {
	case connectivity.TransientFailure:
		return fmt.Errorf("gRPC connection transient failure")
	case connectivity.Shutdown:
		return fmt.Errorf("gRPC connection shutdown")
	default:
		return fmt.Errorf("gRPC connection not ready, state: %s", state)
	}
}

// performHealthCheck performs a standard gRPC health check
func (g *GRPCMonitor) performHealthCheck(ctx context.Context) error {
	// First ensure connection is ready
	if err := g.checkConnectionState(ctx); err != nil {
		return fmt.Errorf("connection not ready for health check: %w", err)
	}

	client := grpc_health_v1.NewHealthClient(g.conn)

	req := &grpc_health_v1.HealthCheckRequest{}
	if g.serviceName != "" {
		req.Service = g.serviceName
	}

	resp, err := client.Check(ctx, req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if resp.Status != grpc_health_v1.HealthCheckResponse_SERVING {
		return fmt.Errorf("service not serving, status: %s", resp.Status)
	}

	return nil
}

// performReflectionCheck performs a gRPC reflection check
func (g *GRPCMonitor) performReflectionCheck(ctx context.Context) error {
	// For now, use a simple connectivity check since reflection API is complex
	// In a real implementation, you would use the reflection API to list services
	// This is a simplified approach that just verifies the server is reachable
	return g.performConnectivityCheck(ctx)
}

// performConnectivityCheck performs a simple connectivity check
func (g *GRPCMonitor) performConnectivityCheck(ctx context.Context) error {
	// For connectivity check, we need to ensure the connection is actually working
	// by attempting a simple operation
	if err := g.checkConnectionState(ctx); err != nil {
		return err
	}

	// Additional verification: try to get connection info
	// This helps ensure the connection is truly established
	state := g.conn.GetState()
	if state != connectivity.Ready {
		return fmt.Errorf("connection not ready after verification, state: %s", state)
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
