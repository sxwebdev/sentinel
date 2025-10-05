package projects

import (
	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store     *store.Store
	validator *validator.Validate
}

func New(store *store.Store) *Service {
	return &Service{store: store, validator: validator.New()}
}
