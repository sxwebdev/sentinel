//nolint:revive,predeclared
package auth

import (
	"golang.org/x/crypto/bcrypt"
)

// IsCorrectPassword checks if the provided password matches the hashed password.
func IsCorrectPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
