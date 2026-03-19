package users

import (
	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store *store.Store

	validator  *validator.Validate
	bcryptCost int
}

// New creates new users service
func New(st *store.Store, bcryptCost int) *Service {
	return &Service{
		store:      st,
		validator:  validator.New(),
		bcryptCost: bcryptCost,
	}
}
