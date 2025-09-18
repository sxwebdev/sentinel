package monitors

import (
	"context"
	"fmt"
	"io"

	"github.com/sxwebdev/sentinel/internal/models"
)

// ServiceMonitor defines the interface for all service monitors
type ServiceMonitor interface {
	io.Closer

	Name() string
	Protocol() models.ServiceProtocolType
	Check(ctx context.Context) error
	Config() *models.Service
}

// NewMonitor creates a new monitor based on the service configuration
func NewMonitor(cfg *models.Service) (ServiceMonitor, error) {
	switch cfg.Protocol {
	case models.ServiceProtocolTypeHTTP:
		return NewHTTPMonitor(cfg)
	case models.ServiceProtocolTypeTCP:
		return NewTCPMonitor(cfg)
	case models.ServiceProtocolTypeGRPC:
		return NewGRPCMonitor(cfg)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}
}

// BaseMonitor provides common functionality for all monitors
type BaseMonitor struct {
	name     string
	protocol models.ServiceProtocolType
	config   *models.Service
}

func NewBaseMonitor(cfg *models.Service) BaseMonitor {
	return BaseMonitor{
		name:     cfg.Name,
		protocol: cfg.Protocol,
		config:   cfg,
	}
}

func (b *BaseMonitor) Name() string {
	return b.name
}

func (b *BaseMonitor) Protocol() models.ServiceProtocolType {
	return b.protocol
}

func (b *BaseMonitor) Config() *models.Service {
	return b.config
}
