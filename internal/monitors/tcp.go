package monitors

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sxwebdev/sentinel/internal/storage"
)

// TCPMonitor monitors TCP endpoints
type TCPMonitor struct {
	BaseMonitor
	sendData   string
	expectData string
}

// NewTCPMonitor creates a new TCP monitor
func NewTCPMonitor(cfg storage.Service) (*TCPMonitor, error) {
	// Extract TCP config
	var tcpConfig *storage.TCPConfig
	if cfg.Config.TCP != nil {
		tcpConfig = cfg.Config.TCP
	}

	monitor := &TCPMonitor{
		BaseMonitor: NewBaseMonitor(cfg),
	}

	// Apply TCP-specific config if available
	if tcpConfig != nil {
		monitor.sendData = tcpConfig.SendData
		monitor.expectData = tcpConfig.ExpectData
	}

	return monitor, nil
}

// Check performs the TCP health check
func (t *TCPMonitor) Check(ctx context.Context) error {
	// Create connection with timeout
	dialer := net.Dialer{
		Timeout: t.config.Timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", t.config.Endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Send data if specified
	if t.sendData != "" {
		_, err = conn.Write([]byte(t.sendData))
		if err != nil {
			return fmt.Errorf("failed to send data: %w", err)
		}
	}

	// Expect data if specified
	if t.expectData != "" {
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
		if !strings.Contains(response, t.expectData) {
			return fmt.Errorf("expected data not found in response: %s", t.expectData)
		}
	}

	return nil
}
