package config

import (
	"time"
)

// Config represents the main configuration structure
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Database   DatabaseConfig   `yaml:"database"`
	Telegram   TelegramConfig   `yaml:"telegram"`
	Services   []ServiceConfig  `yaml:"services"`
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

// TelegramConfig holds Telegram bot configuration
type TelegramConfig struct {
	BotToken string `yaml:"bot_token"`
	ChatID   string `yaml:"chat_id"`
	Enabled  bool   `yaml:"enabled"`
}

// ServiceConfig represents configuration for a single service
type ServiceConfig struct {
	Name     string                 `yaml:"name"`
	Protocol string                 `yaml:"protocol"`
	Endpoint string                 `yaml:"endpoint"`
	Interval time.Duration          `yaml:"interval"`
	Timeout  time.Duration          `yaml:"timeout"`
	Retries  int                    `yaml:"retries"`
	Tags     []string               `yaml:"tags"`
	Config   map[string]interface{} `yaml:"config"`
}

// ServiceStatus represents the current status of a service
type ServiceStatus int

const (
	StatusUnknown ServiceStatus = iota
	StatusUp
	StatusDown
	StatusMaintenance
)

func (s ServiceStatus) String() string {
	switch s {
	case StatusUp:
		return "UP"
	case StatusDown:
		return "DOWN"
	case StatusMaintenance:
		return "MAINTENANCE"
	default:
		return "UNKNOWN"
	}
}

// ServiceState holds the current state of a monitored service
type ServiceState struct {
	Name               string        `json:"name"`
	Protocol           string        `json:"protocol"`
	Endpoint           string        `json:"endpoint"`
	Status             ServiceStatus `json:"status"`
	LastCheck          time.Time     `json:"last_check"`
	NextCheck          time.Time     `json:"next_check"`
	LastError          string        `json:"last_error,omitempty"`
	ConsecutiveFails   int           `json:"consecutive_fails"`
	ConsecutiveSuccess int           `json:"consecutive_success"`
	TotalChecks        int           `json:"total_checks"`
	ResponseTime       time.Duration `json:"response_time"`
	Tags               []string      `json:"tags"`
}

// Incident represents a service incident
type Incident struct {
	ID          string         `json:"id"`
	ServiceName string         `json:"service_name"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     *time.Time     `json:"end_time,omitempty"`
	Error       string         `json:"error"`
	Duration    *time.Duration `json:"duration,omitempty"`
	Resolved    bool           `json:"resolved"`
}
