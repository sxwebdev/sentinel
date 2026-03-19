package hutils

import "strings"

// NormalizeEmail normalizes an email address by trimming whitespace and converting it to lowercase.
func NormalizeEmail(email string) string {
	if email == "" {
		return ""
	}
	// Trim whitespace and convert to lowercase
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}
