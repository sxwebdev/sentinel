package agent

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sxwebdev/sentinel/internal/config"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/tkcrm/mx/logger"
)

type Agent struct {
	logger logger.Logger

	config     *config.ConfigAgent
	systemInfo models.SystemInfo

	token       string
	fingerprint string

	client *client

	mu            sync.RWMutex
	services      []*models.Service
	currentState  ConnectionState
	changeStateCh chan ConnectionState
	isConnection  atomic.Bool
}

// New creates a new Agent instance
func New(
	ctx context.Context,
	l logger.Logger,
	config *config.ConfigAgent,
	systemInfo models.SystemInfo,
) (*Agent, error) {
	a := &Agent{
		logger:        l,
		config:        config,
		systemInfo:    systemInfo,
		changeStateCh: make(chan ConnectionState, 1),
	}

	a.token = config.Token
	a.fingerprint = a.getFingerprint()

	var err error
	a.client, err = newClient(
		ctx,
		a.logger,
		a.token,
		a.fingerprint,
		a.systemInfo,
		config.HubServer,
		a.changeStateCh,
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

// setState sets the connection state
func (s *Agent) setState(state ConnectionState) {
	s.mu.Lock()
	s.currentState = state
	s.mu.Unlock()
}

// getState returns the current connection state
func (s *Agent) getState() ConnectionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentState
}

// setServices sets the services fetched from the hub server
func (s *Agent) setServices(services []*models.Service) {
	s.mu.Lock()
	s.services = services
	s.mu.Unlock()
}
