package storecmn

import "errors"

var (
	ErrEmptyID       = errors.New("empty id")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)
