package models

import (
	"fmt"

	"github.com/sxwebdev/sentinel/pkg/rbacconnect"
)

// GetID returns the ID of the user
func (u User) GetID() string {
	return u.ID
}

// GetEmail returns the email of the user
func (u User) GetEmail() string {
	return u.Email
}

// GetPassword returns the password of the user
func (u User) GetPassword() string {
	return u.Password
}

// GetRole returns the role of the user
func (u User) GetRole() string {
	return u.Role
}

// HasRole checks if the user has the specified role
func (u User) HasRole(role UserRole) bool {
	return u.Role == role
}

type UserRole = rbacconnect.Role

const (
	// UserRoleRoot represents a root user
	UserRoleRoot UserRole = "root"
	// UserRoleAdmin represents an admin user
	UserRoleAdmin UserRole = "admin"
	// UserRoleUser represents a regular user
	UserRoleUser UserRole = "user"
)

// ValidateUserRole checks if the UserRole is valid
func ValidateUserRole(r UserRole) error {
	switch r {
	case UserRoleRoot, UserRoleAdmin, UserRoleUser:
		return nil
	default:
		return fmt.Errorf("invalid user role: %s", r)
	}
}
