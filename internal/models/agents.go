package models

type AgentStatusType string

const (
	AgentStatusTypeUnknown  AgentStatusType = "unknown"
	AgentStatusTypeActive   AgentStatusType = "active"
	AgentStatusTypeInactive AgentStatusType = "inactive"
)

func (s AgentStatusType) String() string {
	return string(s)
}
