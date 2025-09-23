package config

import (
	"fmt"
	"time"

	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

// ConfigHub represents the main configuration structure
type ConfigHub struct {
	Log        logger.Config
	Ops        ops.Config
	DataDir    string           `yaml:"data_dir" default:"./data"`
	Server     ServerConfig     `yaml:"server"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	// Database      DatabaseConfig      `yaml:"database"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Timezone      string              `yaml:"timezone" default:"UTC"`
	Upgrader      Upgrader            `yaml:"upgrader"`
}

// ServerConfig holds web server configuration
type ServerConfig struct {
	Port     int            `yaml:"port" default:"8080"`
	Host     string         `yaml:"host" default:"0.0.0.0"`
	BaseHost string         `yaml:"base_host" default:"localhost:8080"`
	Frontend FrontendConfig `yaml:"frontend"`
	Auth     AuthConfig     `yaml:"auth"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	Enabled bool       `yaml:"enabled"`
	Users   []UserAuth `yaml:"users"`
}

// UserAuth represents a user with basic auth credentials
type UserAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// FrontendConfig holds frontend-specific configuration
type FrontendConfig struct {
	BaseURL   string `yaml:"base_url" default:"http://localhost:8080/api/v1"`
	SocketURL string `yaml:"socket_url" default:"ws://localhost:8080/ws"`
}

// MonitoringConfig holds global monitoring settings
type MonitoringConfig struct {
	Global GlobalConfig `yaml:"global"`
}

// GlobalConfig holds default monitoring parameters
type GlobalConfig struct {
	DefaultInterval time.Duration `yaml:"default_interval" default:"1m"`
	DefaultTimeout  time.Duration `yaml:"default_timeout" default:"10s"`
	DefaultRetries  int64         `yaml:"default_retries" default:"10"`
}

// NotificationsConfig holds notification settings for multiple providers
type NotificationsConfig struct {
	Enabled bool     `yaml:"enabled"`
	URLs    []string `yaml:"urls"`
}

type Upgrader struct {
	IsEnabled bool   `yaml:"is_enabled"`
	Command   string `yaml:"command"`
}

// SetDefaults applies default values to configuration
func (c *ConfigHub) SetDefaults() error {
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

	return nil
}

// Validate checks if configuration is valid
func (c *ConfigHub) Validate() error {
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
