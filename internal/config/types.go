package config

import (
	"time"
)

// Config represents the main configuration structure
type Config struct {
	Server        ServerConfig        `yaml:"server"`
	Monitoring    MonitoringConfig    `yaml:"monitoring"`
	Database      DatabaseConfig      `yaml:"database"`
	Notifications NotificationsConfig `yaml:"notifications"`
}

// ServerConfig holds web server configuration
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
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
