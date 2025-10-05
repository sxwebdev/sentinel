package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type Tags []string

// Scan implements the sql.Scanner interface for
func (t *Tags) Scan(value any) error {
	if value == nil {
		*t = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		if v == "" {
			*t = Tags{}
			return nil
		}
		*t = Tags{}
		return json.Unmarshal([]byte(v), t)
	case []byte:
		if len(v) == 0 {
			*t = Tags{}
			return nil
		}
		*t = Tags{}
		return json.Unmarshal(v, t)
	default:
		return fmt.Errorf("cannot scan %T into Tags", value)
	}
}

// Value implements the driver.Valuer interface for Tags
func (t Tags) Value() (driver.Value, error) {
	if len(t) == 0 {
		return "[]", nil
	}
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal Tags: %w", err)
	}
	return string(bytes), nil
}
