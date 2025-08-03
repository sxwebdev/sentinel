package config

import (
	"time"

	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

// Config represents the main configuration structure
type Config struct {
	Log           logger.Config
	Ops           ops.Config
	Server        ServerConfig        `yaml:"server"`
	Monitoring    MonitoringConfig    `yaml:"monitoring"`
	Database      DatabaseConfig      `yaml:"database"`
	Notifications NotificationsConfig `yaml:"notifications"`
	Timezone      string              `yaml:"timezone"`
	Upgrader      Upgrader            `yaml:"upgrader"`
}

// ServerConfig holds web server configuration
type ServerConfig struct {
	Port     int            `yaml:"port"`
	Host     string         `yaml:"host"`
	BaseHost string         `yaml:"base_host"`
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
	BaseURL   string `yaml:"base_url"`
	SocketURL string `yaml:"socket_url"`
}

// MonitoringConfig holds global monitoring settings
type MonitoringConfig struct {
	Global GlobalConfig `yaml:"global"`
}

// GlobalConfig holds default monitoring parameters
type GlobalConfig struct {
	DefaultInterval time.Duration `yaml:"default_interval"`
	DefaultTimeout  time.Duration `yaml:"default_timeout"`
	DefaultRetries  int           `yaml:"default_retries"`
}

// DatabaseConfig holds database settings
type DatabaseConfig struct {
	Path string `yaml:"path"`
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
