package utils

import "github.com/oklog/ulid/v2"

// GenerateULID generates a new ULID
func GenerateULID() string {
	return ulid.Make().String()
}
