package models

import (
	"database/sql/driver"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type ResourceKindType string

const (
	ResourceKindTypeUnknown ResourceKindType = "unknown"
	ResourceKindTypeServer  ResourceKindType = "server"
	ResourceKindTypeService ResourceKindType = "service"
)

func (s ResourceKindType) String() string {
	return string(s)
}

// Validate checks if the ResourceKindType is valid
func (s ResourceKindType) Validate() error {
	switch s {
	case ResourceKindTypeServer, ResourceKindTypeService:
		return nil
	default:
		return fmt.Errorf("invalid resource kind: %s", s)
	}
}

type ResourcePayload struct{}

// Scan implements the interface for scanning DB values into struct fields.
func (s *ResourcePayload) Scan(value any) error {
	var jsonField storecmn.JSONField
	if err := jsonField.Scan(value); err != nil {
		return err
	}
	return jsonField.ConvertToAny(s)
}

// Value implements the interface for converting struct fields into DB values.
func (s ResourcePayload) Value() (driver.Value, error) {
	var jsonField storecmn.JSONField
	if err := jsonField.UnmarshalFromAny(s); err != nil {
		return nil, err
	}
	return jsonField.Value()
}
