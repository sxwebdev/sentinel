package models

import "fmt"

type AgentStatusType string

const (
	AgentStatusTypeUnknown  AgentStatusType = "unknown"
	AgentStatusTypeActive   AgentStatusType = "active"
	AgentStatusTypeInactive AgentStatusType = "inactive"
)

func (s AgentStatusType) String() string {
	return string(s)
}

type AgentConfig struct{}

type AgentKindType string

const (
	AgentKindTypeHub   AgentKindType = "hub"
	AgentKindTypeAgent AgentKindType = "agent"
)

func (s AgentKindType) String() string {
	return string(s)
}

// Validate agent kind
func (s AgentKindType) Validate() error {
	switch s {
	case AgentKindTypeHub, AgentKindTypeAgent:
		return nil
	default:
		return fmt.Errorf("invalid agent kind: %s", s)
	}
}
