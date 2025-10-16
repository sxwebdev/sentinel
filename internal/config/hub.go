package config

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

// ConfigHub represents the main configuration structure
type ConfigHub struct {
	Log        logger.Config
	Ops        ops.Config
	DataDir    string       `yaml:"data_dir" validate:"required" default:"./data"`
	Server     ServerConfig `yaml:"server"`
	Timezone   string       `yaml:"timezone" default:"UTC"`
	Updater    Updater      `yaml:"updater"`
	KvDbEngine string       `yaml:"kv_db_engine" default:"inmemory" example:"inmemory, badgerdb" validate:"oneof=inmemory badgerdb"`
}

// HubDataDir returns the data directory for the hub
func (c *ConfigHub) HubDataDir() string {
	return filepath.Join(c.DataDir, "hub")
}

// ServerConfig holds web server configuration
type ServerConfig struct {
	Addr string     `yaml:"addr" default:":8080"`
	Auth AuthConfig `yaml:"auth"`
}

// AuthConfig holds authentication settings
type AuthConfig struct {
	AccessTokenSecretKey  string `json:"access_token_secret_key"`
	RefreshTokenSecretKey string `json:"refresh_token_secret_key"`
}

// UserAuth represents a user with basic auth credentials
type UserAuth struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

type Updater struct {
	IsEnabled bool   `yaml:"is_enabled"`
	Command   string `yaml:"command"`
}

// Validate checks if configuration is valid
func (c *ConfigHub) Validate() error {
	// Validate timezone
	if _, err := time.LoadLocation(c.Timezone); err != nil {
		return fmt.Errorf("invalid timezone: %w", err)
	}

	return nil
}
