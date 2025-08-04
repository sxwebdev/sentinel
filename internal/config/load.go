package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/goccy/go-yaml"
)

// Load reads and parses the configuration file
func Load(path string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		var cfg Config
		if err := cfg.setDefaults(); err != nil {
			return nil, fmt.Errorf("failed to set defaults: %w", err)
		}
		return &cfg, nil
	}

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
	if c.Server.BaseHost == "" {
		c.Server.BaseHost = "localhost:8080"
	}

	// Frontend defaults
	if c.Server.Frontend.BaseURL == "" {
		// Auto-detect protocol based on BaseHost or use HTTP as default
		protocol := "http"
		if c.Server.BaseHost != "localhost:8080" && c.Server.BaseHost != "127.0.0.1:8080" {
			// For production domains, assume HTTPS
			protocol = "https"
		}
		c.Server.Frontend.BaseURL = fmt.Sprintf("%s://%s/api/v1", protocol, c.Server.BaseHost)
	}
	if c.Server.Frontend.SocketURL == "" {
		// Auto-detect protocol based on BaseHost or use WS as default
		protocol := "ws"
		if c.Server.BaseHost != "localhost:8080" && c.Server.BaseHost != "127.0.0.1:8080" {
			// For production domains, assume WSS
			protocol = "wss"
		}
		c.Server.Frontend.SocketURL = fmt.Sprintf("%s://%s/ws", protocol, c.Server.BaseHost)
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
		c.Database.Path = "./data/db.sqlite"
	}

	// Timezone defaults
	if c.Timezone == "" {
		c.Timezone = "UTC"
	}

	return nil
}

// validate checks if configuration is valid
func (c *Config) validate() error {
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

	// Validate timezone
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	return nil
}
