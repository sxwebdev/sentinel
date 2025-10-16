package models

import (
	"database/sql/driver"

	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type ProjectSettings struct {
	MonitorDefaults ProjectMonitorDefaults `json:"monitor_defaults"`
}

type ProjectMonitorDefaults struct {
	Interval int64 `json:"interval" default:"60000"` // in milliseconds
	Timeout  int64 `json:"timeout" default:"10000"`  // in milliseconds
	Retries  int64 `json:"retries" default:"10"`
}

// Scan implements the interface for scanning DB values into struct fields.
func (s *ProjectSettings) Scan(value any) error {
	var jsonField storecmn.JSONField
	if err := jsonField.Scan(value); err != nil {
		return err
	}
	return jsonField.ConvertToAny(s)
}

// Value implements the interface for converting struct fields into DB values.
func (s ProjectSettings) Value() (driver.Value, error) {
	var jsonField storecmn.JSONField
	if err := jsonField.UnmarshalFromAny(s); err != nil {
		return nil, err
	}
	return jsonField.Value()
}
