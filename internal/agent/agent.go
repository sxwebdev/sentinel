package agent

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/tkcrm/mx/logger"
)

type Agent struct {
	logger logger.Logger

	config     *config.ConfigAgent
	systemInfo models.SystemInfo

	// server *agentserver.Server

	token       string
	fingerprint string

	connectionManager *connectionManager
}

// New creates a new Agent instance
func New(
	l logger.Logger,
	config *config.ConfigAgent,
	systemInfo models.SystemInfo,
) (*Agent, error) {
	a := &Agent{
		logger:     l,
		config:     config,
		systemInfo: systemInfo,
	}

	a.token = config.Token
	a.fingerprint = a.getFingerprint()

	var err error
	a.connectionManager, err = newConnectionManager(a)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Name returns the name of the agent
func (a *Agent) Name() string {
	return "sentinel-agent"
}

// Start starts the agent
func (a *Agent) Start(_ context.Context) error {
	return nil
}

// Stop stops the agent
func (a *Agent) Stop(_ context.Context) error {
	return nil
}
