package users

import (
	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store *store.Store

	validator *validator.Validate
}

// New creates new users service
func New(st *store.Store) *Service {
	return &Service{
		store:     st,
		validator: validator.New(),
	}
}
