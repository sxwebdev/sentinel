package monitors

import (
	"context"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/storage"
)

// ServiceMonitor defines the interface for all service monitors
type ServiceMonitor interface {
	Name() string
	Protocol() storage.ServiceProtocolType
	Check(ctx context.Context) error
	Config() storage.Service
}

// NewMonitor creates a new monitor based on the service configuration
func NewMonitor(cfg storage.Service) (ServiceMonitor, error) {
	switch cfg.Protocol {
	case storage.ServiceProtocolTypeHTTP:
		return NewHTTPMonitor(cfg)
	case storage.ServiceProtocolTypeTCP:
		return NewTCPMonitor(cfg)
	case storage.ServiceProtocolTypeGRPC:
		return NewGRPCMonitor(cfg)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}
}

// BaseMonitor provides common functionality for all monitors
type BaseMonitor struct {
	name     string
	protocol storage.ServiceProtocolType
	config   storage.Service
}

func NewBaseMonitor(cfg storage.Service) BaseMonitor {
	return BaseMonitor{
		name:     cfg.Name,
		protocol: cfg.Protocol,
		config:   cfg,
	}
}

func (b *BaseMonitor) Name() string {
	return b.name
}

func (b *BaseMonitor) Protocol() storage.ServiceProtocolType {
	return b.protocol
}

func (b *BaseMonitor) Config() storage.Service {
	return b.config
}
