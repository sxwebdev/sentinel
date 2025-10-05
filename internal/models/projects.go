package models

import (
	"database/sql/driver"

	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

type ProjectSettings struct {
	MonitorDefaults ProjectMonitorDefaults `yaml:"monitor_defaults"`
}

type ProjectMonitorDefaults struct {
	DefaultInterval int64 `yaml:"interval" default:"60000"` // in milliseconds
	DefaultTimeout  int64 `yaml:"timeout" default:"10000"`  // in milliseconds
	DefaultRetries  int64 `yaml:"retries" default:"10"`
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
