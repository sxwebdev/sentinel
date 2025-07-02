package monitors

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sxwebdev/sentinel/internal/config"
)

// TCPMonitor monitors TCP connections
type TCPMonitor struct {
	BaseMonitor
	sendData   string
	expectData string
}

// NewTCPMonitor creates a new TCP monitor
func NewTCPMonitor(cfg config.ServiceConfig) (*TCPMonitor, error) {
	return &TCPMonitor{
		BaseMonitor: NewBaseMonitor(cfg),
		sendData:    getConfigString(cfg.Config, "send_data", ""),
		expectData:  getConfigString(cfg.Config, "expect_data", ""),
	}, nil
}

// Check performs the TCP connection check
func (t *TCPMonitor) Check(ctx context.Context) error {
	// Create dialer with timeout
	dialer := &net.Dialer{
		Timeout: t.config.Timeout,
	}

	// Establish connection
	conn, err := dialer.DialContext(ctx, "tcp", t.config.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Set deadline for the entire operation
	deadline := time.Now().Add(t.config.Timeout)
	if err := conn.SetDeadline(deadline); err != nil {
		return fmt.Errorf("failed to set deadline: %w", err)
	}

	// Send data if specified
	if t.sendData != "" {
		_, err := conn.Write([]byte(t.sendData))
		if err != nil {
			return fmt.Errorf("failed to send data: %w", err)
		}

		// If we expect data back, read it
		if t.expectData != "" {
			buffer := make([]byte, 1024)
			n, err := conn.Read(buffer)
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			response := string(buffer[:n])
			if response != t.expectData {
				return fmt.Errorf("unexpected response: got %q, expected %q", response, t.expectData)
			}
		}
	}

	return nil
}
