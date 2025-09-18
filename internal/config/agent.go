package config

import (
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
	"github.com/tkcrm/mx/transport/connectrpc_transport"
)

type ConfigAgent struct {
	Log     logger.Config
	Ops     ops.Config
	DataDir string                      `yaml:"data_dir" default:"./data-agent"`
	Server  connectrpc_transport.Config `yaml:"server"`
	Token   string                      `yaml:"token" validate:"required"`
}
