package agent

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/agent/agentserver"
	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/tkcrm/mx/logger"
)

type Agent struct {
	logger logger.Logger

	config     *config.ConfigAgent
	systemInfo models.SystemInfo

	server *agentserver.Server
}

// New creates a new Agent instance
func New(
	l logger.Logger,
	config *config.ConfigAgent,
	systemInfo models.SystemInfo,
) *Agent {
	return &Agent{
		logger:     l,
		config:     config,
		systemInfo: systemInfo,
		server:     agentserver.New(systemInfo),
	}
}

// Name returns the name of the agent
func (a *Agent) Name() string {
	return "sentinel-agent"
}

// Start starts the agent
func (a *Agent) Start(_ context.Context) error {
	// Placeholder for starting agent logic
	_ = a.getFingerprint()
	return nil
}

// Stop stops the agent
func (a *Agent) Stop(_ context.Context) error {
	// Placeholder for stopping agent logic
	return nil
}
