package monitors

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sxwebdev/sentinel/internal/storage"
)

// TCPConfig represents TCP monitor configuration
type TCPConfig struct {
	Endpoint   string `json:"endpoint" validate:"required,hostname_port"`
	SendData   string `json:"send_data,omitempty"`
	ExpectData string `json:"expect_data,omitempty"`
}

// TCPMonitor monitors TCP endpoints
type TCPMonitor struct {
	BaseMonitor
	conf TCPConfig
}

// NewTCPMonitor creates a new TCP monitor
func NewTCPMonitor(svc storage.Service) (*TCPMonitor, error) {
	// Extract TCP config
	conf, err := GetConfig[TCPConfig](svc.Config, storage.ServiceProtocolTypeTCP)
	if err != nil {
		return nil, fmt.Errorf("failed to get TCP config: %w", err)
	}

	monitor := &TCPMonitor{
		BaseMonitor: NewBaseMonitor(svc),
		conf:        conf,
	}

	return monitor, nil
}

// Check performs the TCP health check
func (t *TCPMonitor) Check(ctx context.Context) error {
	// Get endpoint from config
	if t.conf.Endpoint == "" {
		return fmt.Errorf("TCP endpoint not configured")
	}
	endpoint := t.conf.Endpoint

	// Create connection with timeout
	dialer := net.Dialer{
		Timeout: t.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Send data if specified
	if t.conf.SendData != "" {
		_, err = conn.Write([]byte(t.conf.SendData))
		if err != nil {
			return fmt.Errorf("failed to send data: %w", err)
		}
	}

	// Expect data if specified
	if t.conf.ExpectData != "" {
		// Set read deadline
		deadline := time.Now().Add(t.config.Timeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}

		// Read response
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		response := string(buffer[:n])
		if !strings.Contains(response, t.conf.ExpectData) {
			return fmt.Errorf("expected data not found in response: %s", t.conf.ExpectData)
		}
	}

	return nil
}

// Close implements io.Closer for TCP monitor (no-op since TCP doesn't maintain persistent connections)
func (t *TCPMonitor) Close() error {
	return nil
}
