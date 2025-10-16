package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type AgentStatusType string

const (
	AgentStatusTypeUnknown  AgentStatusType = "unknown"
	AgentStatusTypeActive   AgentStatusType = "active"
	AgentStatusTypeInactive AgentStatusType = "inactive"
)

func (s AgentStatusType) String() string {
	return string(s)
}

type AgentKindType string

const (
	AgentKindTypeHub      AgentKindType = "hub"
	AgentKindTypeExternal AgentKindType = "external"
)

func (s AgentKindType) String() string {
	return string(s)
}

// Validate agent kind
func (s AgentKindType) Validate() error {
	switch s {
	case AgentKindTypeHub, AgentKindTypeExternal:
		return nil
	default:
		return fmt.Errorf("invalid agent kind: %s", s)
	}
}

type AgentConfig struct{}

// Scan implements the interface for scanning DB values into struct fields.
func (s *AgentConfig) Scan(value any) error {
	var jsonField storecmn.JSONField
	if err := jsonField.Scan(value); err != nil {
		return err
	}
	return jsonField.ConvertToAny(s)
}

// Value implements the interface for converting struct fields into DB values.
func (s AgentConfig) Value() (driver.Value, error) {
	var jsonField storecmn.JSONField
	if err := jsonField.UnmarshalFromAny(s); err != nil {
		return nil, err
	}
	return jsonField.Value()
}
