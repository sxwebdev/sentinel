package config

import (
	"github.com/tkcrm/mx/clients/connectrpc_client"
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

type ConfigAgent struct {
	Log       logger.Config
	Ops       ops.Config
	HubServer connectrpc_client.Config `yaml:"hub_server"`
	DataDir   string                   `yaml:"data_dir" default:"./data-agent"`
	Token     string                   `yaml:"token" validate:"required"`
}
