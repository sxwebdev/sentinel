package monitors

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sxwebdev/sentinel/internal/storage"
	"github.com/sxwebdev/sentinel/internal/utils"
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

	timeout := t.config.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}

	// Create connection with timeout
	dialer := net.Dialer{
		Timeout: timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", endpoint)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer conn.Close()

	// Send data if specified
	if t.conf.SendData != "" {
		if err := conn.SetWriteDeadline(time.Now().Add(timeout)); err != nil {
			return fmt.Errorf("failed to set write deadline: %w", err)
		}

		_, err = conn.Write([]byte(t.conf.SendData))
		if err != nil {
			return fmt.Errorf("failed to send data: %w", err)
		}
	}

	// Expect data if specified
	if t.conf.ExpectData != "" {
		// Set read deadline
		deadline := time.Now().Add(timeout)
		err = conn.SetReadDeadline(deadline)
		if err != nil {
			return fmt.Errorf("failed to set read deadline: %w", err)
		}

		var response strings.Builder
		reader := bufio.NewReader(conn)

		for {
			buffer := make([]byte, 1024)
			n, err := reader.Read(buffer)
			if err != nil {
				if utils.IsErrTimeout(err) {
					break
				}
				return fmt.Errorf("failed to read response: %w", err)
			}

			if n == 0 {
				break
			}

			if _, err := response.Write(buffer[:n]); err != nil {
				return fmt.Errorf("failed to write response: %w", err)
			}

			if reader.Buffered() == 0 {
				break
			}
		}

		// Now process the complete response
		receivedData := response.String()

		if len(receivedData) == 0 {
			return fmt.Errorf("server sent no data, expected: '%s'", t.conf.ExpectData)
		}

		// Check if we have the expected data
		if !strings.Contains(receivedData, t.conf.ExpectData) {
			return fmt.Errorf("expected data '%s' not found in response: '%s'", t.conf.ExpectData, receivedData)
		}
	}

	return nil
}

// Close implements io.Closer for TCP monitor (no-op since TCP doesn't maintain persistent connections)
func (t *TCPMonitor) Close() error {
	return nil
}
