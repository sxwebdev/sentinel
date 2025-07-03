package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables
	expanded := expandEnvVars(string(data))

	var cfg Config
	if err := yaml.Unmarshal([]byte(expanded), &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Apply defaults and validate
	if err := cfg.setDefaults(); err != nil {
		return nil, fmt.Errorf("failed to set defaults: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// expandEnvVars replaces ${VAR} with environment variable values
func expandEnvVars(s string) string {
	re := regexp.MustCompile(`\$\{([^}]+)\}`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		varName := match[2 : len(match)-1] // Remove ${ and }
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Return original if env var not found
	})
}

// setDefaults applies default values to configuration
func (c *Config) setDefaults() error {
	// Server defaults
	if c.Server.Host == "" {
		c.Server.Host = "0.0.0.0"
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}

	// Monitoring defaults
	if c.Monitoring.Global.DefaultInterval == 0 {
		c.Monitoring.Global.DefaultInterval = 30 * time.Second
	}
	if c.Monitoring.Global.DefaultTimeout == 0 {
		c.Monitoring.Global.DefaultTimeout = 10 * time.Second
	}
	if c.Monitoring.Global.DefaultRetries == 0 {
		c.Monitoring.Global.DefaultRetries = 3
	}

	// Database defaults
	if c.Database.Path == "" {
		c.Database.Path = "./data/incidents.db"
	}

	// Service defaults
	for i := range c.Services {
		svc := &c.Services[i]
		if svc.Interval == 0 {
			svc.Interval = c.Monitoring.Global.DefaultInterval
		}
		if svc.Timeout == 0 {
			svc.Timeout = c.Monitoring.Global.DefaultTimeout
		}
		if svc.Retries == 0 {
			svc.Retries = c.Monitoring.Global.DefaultRetries
		}
	}

	return nil
}

// validate checks if configuration is valid
func (c *Config) validate() error {
	if len(c.Services) == 0 {
		return fmt.Errorf("no services configured")
	}

	serviceNames := make(map[string]bool)
	for _, svc := range c.Services {
		if svc.Name == "" {
			return fmt.Errorf("service name cannot be empty")
		}
		if serviceNames[svc.Name] {
			return fmt.Errorf("duplicate service name: %s", svc.Name)
		}
		serviceNames[svc.Name] = true

		if svc.Protocol == "" {
			return fmt.Errorf("service %s: protocol cannot be empty", svc.Name)
		}
		if svc.Endpoint == "" {
			return fmt.Errorf("service %s: endpoint cannot be empty", svc.Name)
		}

		// Validate protocol
		validProtocols := []string{"http", "https", "tcp", "grpc", "redis"}
		valid := false
		for _, p := range validProtocols {
			if strings.EqualFold(svc.Protocol, p) {
				valid = true
				break
			}
		}
		if !valid {
			return fmt.Errorf("service %s: unsupported protocol %s", svc.Name, svc.Protocol)
		}

		if svc.Interval < time.Second {
			return fmt.Errorf("service %s: interval must be at least 1 second", svc.Name)
		}
		if svc.Timeout < time.Second {
			return fmt.Errorf("service %s: timeout must be at least 1 second", svc.Name)
		}
		if svc.Retries < 1 {
			return fmt.Errorf("service %s: retries must be at least 1", svc.Name)
		}
	}

	// Validate notifications config if enabled
	if c.Notifications.Enabled {
		if len(c.Notifications.URLs) == 0 {
			return fmt.Errorf("notification URLs are required when notifications are enabled")
		}

		for i, url := range c.Notifications.URLs {
			if url == "" {
				return fmt.Errorf("notification URL at index %d cannot be empty", i)
			}
		}
	}

	return nil
}
