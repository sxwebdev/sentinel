package models

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
