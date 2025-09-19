package incidents

import (
	"context"

	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
)

// GetByID retrieves an incident by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (*models.Incident, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	return s.store.Incidents().GetByID(ctx, id)
}

// Delete removes an incident by its ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	return s.store.Incidents().Delete(ctx, id)
}

// GetAllUnresolvedByServiceID retrieves all unresolved incidents associated with a specific service ID.
func (s *Service) GetAllUnresolvedByServiceID(ctx context.Context, serviceID string) ([]*models.Incident, error) {
	if serviceID == "" {
		return nil, storecmn.ErrEmptyID
	}

	return s.store.Incidents().GetAllUnresolvedByServiceID(ctx, serviceID)
}

// ResolveByID resolves a specific incident by its ID.
func (s *Service) ResolveByID(ctx context.Context, id string) (*models.Incident, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	if err := s.store.Incidents().ResolveByID(ctx, id); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

type CreateParams struct {
	ServiceID string `db:"service_id" json:"service_id" validate:"required"`
	Error     string `db:"error" json:"error" validate:"required"`
}

// Create creates a new incident.
func (s *Service) Create(ctx context.Context, params CreateParams) (*models.Incident, error) {
	if err := validator.New().Struct(params); err != nil {
		return nil, err
	}

	return s.store.Incidents().Create(ctx, utils.GenerateULID(), params.ServiceID, params.Error)
}
