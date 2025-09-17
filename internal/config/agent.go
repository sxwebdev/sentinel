package config

import (
	"fmt"

	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
	"github.com/tkcrm/mx/transport/connectrpc_transport"
)

type ConfigAgent struct {
	Log    logger.Config
	Ops    ops.Config
	Server connectrpc_transport.Config `yaml:"server"`
	Token  string                      `yaml:"token"`
}

func (c *ConfigAgent) SetDefaults() error {
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Server.Addr == "" {
		c.Server.Addr = "localhost:9000"
	}
	return nil
}

func (c *ConfigAgent) Validate() error {
	if c.Token == "" {
		return fmt.Errorf("token is required")
	}
	return nil
}
