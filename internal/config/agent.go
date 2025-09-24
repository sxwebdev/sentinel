package config

import (
	"path/filepath"

	"github.com/tkcrm/mx/clients/connectrpc_client"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

type ConfigAgent struct {
	Log       logger.Config
	Ops       ops.Config
	HubServer connectrpc_client.Config `yaml:"hub_server"`
	DataDir   string                   `yaml:"data_dir" validate:"required" default:"./data"`
	Token     string                   `yaml:"token" validate:"required"`
}

// AgentDataDir returns the data directory for the agent
func (c *ConfigAgent) AgentDataDir() string {
	return filepath.Join(c.DataDir, "agent")
}
