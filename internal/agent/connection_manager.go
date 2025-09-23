package agent

// ConnectionState represents the state of the connection.
type ConnectionState uint8

const (
	ConnectionStateDisconnected ConnectionState = iota // No active connection
	ConnectionStateConnected                           // Connected
)

type connectionManager struct {
	agent *Agent
	state ConnectionState
}

// newConnectionManager creates a new instance of connectionManager.
func newConnectionManager(agent *Agent) *connectionManager {
	return &connectionManager{
		agent: agent,
		state: ConnectionStateDisconnected,
	}
}

// isConnected checks if the connection is active.
func (cm *connectionManager) isConnected() bool {
	return cm.state == ConnectionStateConnected
}
