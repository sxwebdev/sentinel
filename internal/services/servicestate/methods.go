package servicestate

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
)

// GetAll retrieves all service states
func (s *Service) GetAll(ctx context.Context) ([]*models.ServiceState, error) {
	return s.store.ServiceStates().GetAll(ctx)
}

// GetByServiceID retrieves a service state by service ID
func (s *Service) GetByServiceID(ctx context.Context, serviceID string) (*models.ServiceState, error) {
	return s.store.ServiceStates().GetByServiceID(ctx, serviceID)
}

type UpdateParams = repo_service_states.UpdateRequest

// Update updates a service state by ID
func (s *Service) Update(ctx context.Context, id string, params UpdateParams) (*models.ServiceState, error) {
	if err := params.Status.Validate(); err != nil {
		return nil, err
	}
	return s.store.ServiceStates().Update(ctx, id, params)
}

// Stats represents aggregated statistics about service states
func (s *Service) Stats(ctx context.Context) (*repo_service_states.StatsRow, error) {
	return s.store.ServiceStates().Stats(ctx)
}
