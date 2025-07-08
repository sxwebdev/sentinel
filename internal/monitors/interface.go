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

// getConfigString safely gets a string value from config map
func getConfigString(cfg map[string]any, key string, defaultValue string) string {
	if cfg == nil {
		return defaultValue
	}
	if val, ok := cfg[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getConfigInt safely gets an int value from config map
func getConfigInt(cfg map[string]any, key string, defaultValue int) int {
	if cfg == nil {
		return defaultValue
	}
	if val, ok := cfg[key]; ok {
		switch v := val.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// getConfigBool safely gets a bool value from config map
func getConfigBool(cfg map[string]any, key string, defaultValue bool) bool {
	if cfg == nil {
		return defaultValue
	}
	if val, ok := cfg[key]; ok {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return defaultValue
}

// getConfigHeaders safely gets headers from config map
func getConfigHeaders(cfg map[string]any) map[string]string {
	headers := make(map[string]string)
	if cfg == nil {
		return headers
	}

	if val, ok := cfg["headers"]; ok {
		if headerMap, ok := val.(map[string]any); ok {
			for k, v := range headerMap {
				if str, ok := v.(string); ok {
					headers[k] = str
				}
			}
		}
	}
	return headers
}
