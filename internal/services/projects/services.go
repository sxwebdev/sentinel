package projects

import (
	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/store"
)

type Service struct {
	store     *store.Store
	validator *validator.Validate

	agentsService *agents.Service
}

func New(store *store.Store, agentsService *agents.Service) *Service {
	return &Service{
		store:         store,
		validator:     validator.New(),
		agentsService: agentsService,
	}
}
