package monitors

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
)

// ServiceMonitor defines the interface for all service monitors
type ServiceMonitor interface {
	io.Closer
	Check(ctx context.Context) error
}

type MonitorParams struct {
	ServiceName string
	Protocol    models.ServiceProtocolType
	Timeout     time.Duration
	Config      map[string]any
}

// NewMonitor creates a new monitor based on the service configuration
func NewMonitor(params MonitorParams) (ServiceMonitor, error) {
	switch params.Protocol {
	case models.ServiceProtocolTypeHTTP:
		return newHTTPMonitor(params)
	case models.ServiceProtocolTypeTCP:
		return newTCPMonitor(params)
	case models.ServiceProtocolTypeGRPC:
		return newGRPCMonitor(params)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", params.Protocol)
	}
}

// baseMonitor provides common functionality for all monitors
type baseMonitor struct {
	name     string
	protocol models.ServiceProtocolType
	timeout  time.Duration
}

func newBaseMonitor(name string, protocol models.ServiceProtocolType, timeout time.Duration) baseMonitor {
	return baseMonitor{
		name:     name,
		protocol: protocol,
		timeout:  timeout,
	}
}
