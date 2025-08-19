package web

import "errors"

var (
	ErrServiceIDRequired   = errors.New("service ID is required")
	ErrServiceNameRequired = errors.New("service name is required")
	ErrProtocolRequired    = errors.New("protocol is required")
	ErrIncidentIDRequired  = errors.New("incident ID is required")
)
