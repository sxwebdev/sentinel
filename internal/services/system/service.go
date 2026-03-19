package system

import (
	"github.com/sxwebdev/sentinel/internal/services/users"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store *store.Store

	usersService *users.Service
}

// New creates new system service
func New(st *store.Store, usersService *users.Service) *Service {
	return &Service{
		store:        st,
		usersService: usersService,
	}
}
