package monitors

import (
	"context"
	"fmt"
	"strings"

	"github.com/sxwebdev/sentinel/internal/storage"
)

// ServiceMonitor defines the interface for all service monitors
type ServiceMonitor interface {
	Name() string
	Protocol() string
	Check(ctx context.Context) error
	Config() storage.Service
}

// NewMonitor creates a new monitor based on the service configuration
func NewMonitor(cfg storage.Service) (ServiceMonitor, error) {
	protocol := strings.ToLower(cfg.Protocol)

	switch protocol {
	case "http", "https":
		return NewHTTPMonitor(cfg)
	case "tcp":
		return NewTCPMonitor(cfg)
	case "grpc":
		return NewGRPCMonitor(cfg)
	case "redis":
		return NewRedisMonitor(cfg)
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", cfg.Protocol)
	}
}

// BaseMonitor provides common functionality for all monitors
type BaseMonitor struct {
	name     string
	protocol string
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

func (b *BaseMonitor) Protocol() string {
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
