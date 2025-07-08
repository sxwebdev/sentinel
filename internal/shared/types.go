package shared

import "time"

// MultiEndpointConfig represents configuration for monitoring multiple endpoints
type MultiEndpointConfig struct {
	Endpoints []EndpointConfig `json:"endpoints" yaml:"endpoints"`
	Condition string           `json:"condition" yaml:"condition"` // JavaScript condition
	Timeout   time.Duration    `json:"timeout" yaml:"timeout"`
}

// EndpointConfig represents a single endpoint configuration
type EndpointConfig struct {
	Name     string            `json:"name" yaml:"name"`
	URL      string            `json:"url" yaml:"url"`
	Method   string            `json:"method" yaml:"method"`
	Headers  map[string]string `json:"headers" yaml:"headers"`
	Body     string            `json:"body" yaml:"body"`
	JSONPath string            `json:"json_path" yaml:"json_path"` // Path to extract value from JSON response
	Username string            `json:"username" yaml:"username"`   // Basic Auth username
	Password string            `json:"password" yaml:"password"`   // Basic Auth password
}
