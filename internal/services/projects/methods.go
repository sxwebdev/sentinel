package projects

import (
	"context"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_projects"
	"github.com/sxwebdev/sentinel/internal/utils"
)

// GetAll retrieves all projects from the store.
func (s *Service) GetAll(ctx context.Context) ([]*models.Project, error) {
	return s.store.Projects().GetAll(ctx)
}

// GetByID retrieves a project by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (*models.Project, error) {
	return s.store.Projects().GetByID(ctx, id)
}

type CreateParams = repo_projects.CreateParams

// Create adds a new project to the store.
func (s *Service) Create(ctx context.Context, params CreateParams) (*models.Project, error) {
	params.ID = utils.GenerateULID()

	if err := s.validator.Struct(params); err != nil {
		return nil, err
	}

	if params.Settings.MonitorDefaults.Interval == 0 {
		params.Settings.MonitorDefaults.Interval = 60000
	}

	if params.Settings.MonitorDefaults.Timeout == 0 {
		params.Settings.MonitorDefaults.Timeout = 10000
	}

	if params.Settings.MonitorDefaults.Retries == 0 {
		params.Settings.MonitorDefaults.Retries = 10
	}

	return s.store.Projects().Create(ctx, params)
}

type UpdateParams = repo_projects.UpdateParams

// Update modifies an existing project in the store.
func (s *Service) Update(ctx context.Context, params UpdateParams) (*models.Project, error) {
	if err := s.validator.Struct(params); err != nil {
		return nil, err
	}

	if params.Settings.MonitorDefaults.Interval == 0 {
		params.Settings.MonitorDefaults.Interval = 60000
	}

	if params.Settings.MonitorDefaults.Timeout == 0 {
		params.Settings.MonitorDefaults.Timeout = 10000
	}

	if params.Settings.MonitorDefaults.Retries == 0 {
		params.Settings.MonitorDefaults.Retries = 10
	}

	return s.store.Projects().Update(ctx, params)
}

// Delete removes a project from the store by its ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.store.Projects().Delete(ctx, id)
}
