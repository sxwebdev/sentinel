package agent

import (
	"context"
	"sync"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/hub/hubclient"
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

	state ConnectionState

	client *hubclient.Client

	mu       sync.RWMutex
	services []*models.Service
}

// New creates a new Agent instance
func New(
	ctx context.Context,
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
	a.client, err = hubclient.New(
		ctx,
		a.logger,
		a.token,
		a.fingerprint,
		a.systemInfo,
		config.HubServer,
	)
	if err != nil {
		return nil, err
	}

	return a, nil
}

// Name returns the name of the agent
func (s *Agent) Name() string {
	return "sentinel-agent"
}

// Start starts the agent
func (s *Agent) Start(ctx context.Context) error {
	go s.initConnection(ctx)

	return nil
}

// Stop stops the agent
func (s *Agent) Stop(_ context.Context) error {
	return nil
}

// setServices sets the services fetched from the hub server
func (s *Agent) setServices(services []*models.Service) {
	s.mu.Lock()
	s.services = services
	s.mu.Unlock()
}
