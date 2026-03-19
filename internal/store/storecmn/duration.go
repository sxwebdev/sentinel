package storecmn

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

type Duration time.Duration

// UnmarshalJSON implements the json.Unmarshaler interface.
func (d *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	duration, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(duration)
	return nil
}

// MarshalJSON implements the json.Marshaler interface.
func (d Duration) MarshalJSON() ([]byte, error) {
	s := time.Duration(d).String()
	return json.Marshal(s)
}

// Value implements the driver.Valuer interface.
func (d Duration) Value() (driver.Value, error) {
	return time.Duration(d).String(), nil
}

// Scan implements the sql.Scanner interface.
func (d *Duration) Scan(value any) error {
	if v, ok := value.(string); ok {
		duration, err := time.ParseDuration(v)
		if err != nil {
			return err
		}
		*d = Duration(duration)
		return nil
	}
	return fmt.Errorf("unsupported scan type: %T", value)
}

// String returns the string representation of the duration
func (d Duration) String() string {
	return time.Duration(d).String()
}

// ToDuration converts Duration to time.Duration
func (d Duration) ToDuration() time.Duration {
	return time.Duration(d)
}
