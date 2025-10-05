package agent

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sxwebdev/sentinel/internal/checker"
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

	client  *client
	checker *checker.Checker
	// receiver       *receiver.Receiver
	checkResultsCh chan checkResult

	mu sync.RWMutex
	// services      []*models.Service
	currentState  connectionState
	changeStateCh chan connectionState
	isConnection  atomic.Bool
}

// New creates a new Agent instance
func New(
	ctx context.Context,
	l logger.Logger,
	config *config.ConfigAgent,
	// receiver *receiver.Receiver,
	systemInfo models.SystemInfo,
) (*Agent, error) {
	a := &Agent{
		logger: l,
		config: config,
		// receiver:       receiver,
		systemInfo:     systemInfo,
		changeStateCh:  make(chan connectionState, 1),
		checkResultsCh: make(chan checkResult, 100),
	}

	a.token = config.Token
	a.fingerprint = a.getFingerprint()

	a.checker = checker.New(
		l,
		nil,
		a.onSuccess,
		a.onFailure,
	)

	var err error
	a.client, err = newClient(
		ctx,
		a.logger,
		a.token,
		a.fingerprint,
		a.systemInfo,
		config.HubServer,
		a.changeStateCh,
		a.checkResultsCh,
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

	errCh := make(chan error, 1)
	go func() {
		errCh <- s.checker.Start(ctx)
	}()

	return <-errCh
}

// Stop stops the agent
func (s *Agent) Stop(ctx context.Context) error {
	return s.checker.Stop(ctx)
}

// setState sets the connection state
func (s *Agent) setState(state connectionState) {
	s.mu.Lock()
	s.currentState = state
	s.mu.Unlock()
}

// getState returns the current connection state
func (s *Agent) getState() connectionState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentState
}

// setServices sets the services fetched from the hub server
// func (s *Agent) setServices(services []*models.Service) {
// 	s.mu.Lock()
// 	s.services = services
// 	s.mu.Unlock()
// }
