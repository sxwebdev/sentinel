package monitors

import (
	"encoding/json"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/storage"
)

type Config struct {
	HTTP *HTTPConfig `json:"http,omitempty"`
	TCP  *TCPConfig  `json:"tcp,omitempty"`
	GRPC *GRPCConfig `json:"grpc,omitempty"`
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
