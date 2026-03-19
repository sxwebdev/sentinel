package storecmn

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONField json.RawMessage

// Mashals JSONField to a JSON string
func (j JSONField) MarshalJSON() ([]byte, error) {
	if len(j) == 0 {
		return []byte("null"), nil
	}
	return j, nil
}

// Unmarshals a JSON string to JSONField
func (j *JSONField) UnmarshalJSON(data []byte) error {
	if j == nil {
		return fmt.Errorf("JSONField: UnmarshalJSON on nil pointer")
	}
	*j = append((*j)[0:0], data...)
	return nil
}

// Unmarshal any to JSONField
func (j *JSONField) UnmarshalFromAny(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	switch v := value.(type) {
	case string:
		*j = JSONField(v)
		return nil
	case []byte:
		*j = JSONField(v)
		return nil
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal value: %w", err)
		}
		*j = JSONField(bytes)
	}
	return nil
}

// Scan implements the sql.Scanner interface for JSONField
func (j *JSONField) Scan(value any) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	default:
		return fmt.Errorf("cannot scan %T into JSONField", value)
	}

	// Skip empty strings or whitespace-only values
	trimmed := []byte{}
	for _, b := range data {
		if b > 32 { // Skip control characters and spaces
			trimmed = append(trimmed, b)
		}
	}

	if len(trimmed) == 0 {
		*j = nil
		return nil
	}

	// Validate JSON before assigning
	if !json.Valid(trimmed) {
		// If invalid, try to recover or log and return error
		return fmt.Errorf("invalid JSON data in database: %s", string(data))
	}

	*j = JSONField(trimmed)
	return nil
}

// Value implements the driver.Valuer interface for JSONField
func (j JSONField) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}

	// Validate that the JSON is valid before storing
	if !json.Valid(j) {
		return nil, fmt.Errorf("invalid JSON data")
	}

	return string(j), nil
}

// ConvertToMap converts JSONField to map[string]any
func (j JSONField) ConvertToMap() map[string]any {
	if len(j) == 0 {
		return map[string]any{}
	}
	var result map[string]any
	if err := json.Unmarshal(j, &result); err != nil {
		return map[string]any{}
	}
	return result
}

// ConvertToAny converts JSONField to any
func (j JSONField) ConvertToAny(dst any) error {
	if len(j) == 0 {
		return nil
	}
	return json.Unmarshal(j, dst)
}
