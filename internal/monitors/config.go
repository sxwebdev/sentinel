package monitors

import (
	"encoding/json"
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/storage"
)

type Config struct {
	HTTP *HTTPConfig `json:"http,omitempty"`
	TCP  *TCPConfig  `json:"tcp,omitempty"`
	GRPC *GRPCConfig `json:"grpc,omitempty"`
}

// convertFlatConfigToMonitorConfig converts JSON config object to proper MonitorConfig structure
func (s *Config) Validate(protocol storage.ServiceProtocolType) error {
	v := validator.New(validator.WithRequiredStructEnabled())

	// Validate and convert based on protocol
	switch protocol {
	case storage.ServiceProtocolTypeHTTP:
		if s.HTTP == nil {
			return fmt.Errorf("HTTP config is required for HTTP protocol")
		}

		// Validate HTTP config
		if err := v.Struct(s.HTTP); err != nil {
			return fmt.Errorf("invalid HTTP config: %w", err)
		}

		return nil
	case storage.ServiceProtocolTypeTCP:
		if s.TCP == nil {
			return fmt.Errorf("TCP config is required for TCP protocol")
		}

		// Validate TCP config
		if err := v.Struct(s.TCP); err != nil {
			return fmt.Errorf("invalid TCP config: %w", err)
		}

		return nil
	case storage.ServiceProtocolTypeGRPC:
		if s.GRPC == nil {
			return fmt.Errorf("gRPC config is required for gRPC protocol")
		}

		// Validate gRPC config
		if err := v.Struct(s.GRPC); err != nil {
			return fmt.Errorf("invalid gRPC config: %w", err)
		}

		return nil
	default:
		return fmt.Errorf("unsupported protocol: %s", protocol)
	}
}

func GetConfig[T any](cfg map[string]any, protocol storage.ServiceProtocolType) (T, error) {
	var c T

	if cfg == nil {
		return c, fmt.Errorf("config is nil")
	}

	data, ok := cfg[string(protocol)]
	if !ok {
		return c, fmt.Errorf("config not found for protocol: %s", protocol)
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(jsonData, &c); err != nil {
		return c, err
	}
	return c, nil
}

// ConvertToMap converts the config to a map[string]any
func (c *Config) ConvertToMap() map[string]any {
	return map[string]any{
		string(storage.ServiceProtocolTypeHTTP): c.HTTP,
		string(storage.ServiceProtocolTypeTCP):  c.TCP,
		string(storage.ServiceProtocolTypeGRPC): c.GRPC,
	}
}

// ConvertFromMap converts a map[string]any to Config
func ConvertFromMap(cfg map[string]any) (Config, error) {
	conf := Config{}
	if cfg == nil {
		return conf, fmt.Errorf("config is nil")
	}

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return conf, err
	}

	if err := json.Unmarshal(jsonData, &conf); err != nil {
		return conf, err
	}

	return conf, nil
}
