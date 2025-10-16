package storecmn

import "errors"

var (
	ErrEmptyID         = errors.New("empty id")
	ErrNotFound        = errors.New("not found")
	ErrAlreadyExists   = errors.New("already exists")
	ErrUserNotFound    = errors.New("user not found")
	ErrProjectNotFound = errors.New("project not found")
	ErrEmptyProjectID  = errors.New("empty project id")
)
