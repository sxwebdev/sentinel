package incidents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_incidents"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
)

// GetByID retrieves an incident by its ID.
func (s *Service) GetByID(ctx context.Context, id string) (*models.Incident, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	item, err := s.store.Incidents().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storecmn.ErrNotFound
		}
		return nil, err
	}

	return item, nil
}

// Delete removes an incident by its ID.
func (s *Service) Delete(ctx context.Context, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	return s.store.Incidents().Delete(ctx, id)
}

// GetAllUnresolvedByServiceID retrieves all unresolved incidents associated with a specific service ID.
// func (s *Service) GetAllUnresolvedByServiceID(ctx context.Context, serviceID string) ([]*models.Incident, error) {
// 	if serviceID == "" {
// 		return nil, storecmn.ErrEmptyID
// 	}

// 	return s.store.Incidents().GetAllUnresolvedByMonitorID(ctx, serviceID)
// }

// ResolveByID resolves a specific incident by its ID.
func (s *Service) ResolveByID(ctx context.Context, id string, opts ...repos.Option) (*models.Incident, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	if err := s.store.Incidents(opts...).ResolveByID(ctx, id); err != nil {
		return nil, err
	}

	item, err := s.store.Incidents().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return item, nil
}

type CreateParams struct {
	MonitorID string `validate:"required"`
	Error     string `validate:"required"`
}

// Create creates a new incident.
func (s *Service) Create(ctx context.Context, params CreateParams) (*models.Incident, error) {
	if err := validator.New().Struct(params); err != nil {
		return nil, err
	}

	return s.store.Incidents().Create(ctx, repo_incidents.CreateParams{
		ID:        utils.GenerateULID(),
		MonitorID: &params.MonitorID,
		Summary:   params.Error,
	})
}

type FindParams = repo_incidents.FindParams

// Find retrieves a list of incidents based on the provided filtering parameters.
func (s *Service) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Incident], error) {
	return s.store.Incidents().Find(ctx, params)
}

// Stats retrieves aggregated statistics about incidents.
func (s *Service) Stats(ctx context.Context) (*repo_incidents.StatsRow, error) {
	return s.store.Incidents().Stats(ctx)
}

// StatsByServiceID retrieves statistics about incidents for a specific service within a given time frame.
func (s *Service) StatsByServiceID(ctx context.Context, monitorID string, startTime time.Time) (*repo_incidents.StatsByMonitorIDRow, error) {
	if monitorID == "" {
		return nil, storecmn.ErrEmptyID
	}

	if startTime.IsZero() {
		return nil, fmt.Errorf("start time is required")
	}

	return s.store.Incidents().StatsByMonitorID(ctx, &monitorID, startTime)
}
