package dbutils

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type JSONField json.RawMessage

func (j *JSONField) Scan(value any) error {
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
		return fmt.Errorf("cannot scan %T into JSONField", value)
	}
}

func (j JSONField) Value() (driver.Value, error) {
	if len(j) == 0 {
		return nil, nil
	}
	return string(j), nil
}
