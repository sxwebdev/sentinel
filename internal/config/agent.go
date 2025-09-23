package config

import (
	"github.com/tkcrm/mx/logger"
	"github.com/tkcrm/mx/ops"
)

type ConfigAgent struct {
	Log     logger.Config
	Ops     ops.Config
	DataDir string `yaml:"data_dir" default:"./data-agent"`
	Token   string `yaml:"token" validate:"required"`
}
