package agent

import "github.com/sxwebdev/sentinel/internal/hub/hubclient"

// ConnectionState represents the state of the connection.
type ConnectionState uint8

const (
	ConnectionStateDisconnected ConnectionState = iota // No active connection
	ConnectionStateConnected                           // Connected
)

type connectionManager struct {
	agent *Agent
	state ConnectionState

	client *hubclient.Client
}

// newConnectionManager creates a new instance of connectionManager.
func newConnectionManager(agent *Agent) (*connectionManager, error) {
	client, err := hubclient.New(
		agent.logger,
		agent.fingerprint,
		agent.systemInfo,
		agent.config.HubServer,
	)
	if err != nil {
		return nil, err
	}

	return &connectionManager{
		agent:  agent,
		state:  ConnectionStateDisconnected,
		client: client,
	}, nil
}

// isConnected checks if the connection is active.
func (cm *connectionManager) isConnected() bool {
	return cm.state == ConnectionStateConnected
}
