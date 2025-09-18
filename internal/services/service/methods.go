package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/store/repos"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_services"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/sxwebdev/sentinel/pkg/dbutils"
)

type CreateParams = repo_services.CreateParams

// Create new service
func (s *Service) Create(ctx context.Context, params CreateParams) (*models.ServiceFullView, error) {
	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	params.ID = utils.GenerateULID()

	err := dbutils.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
		// Create service
		_, err := s.store.Services(repos.WithTx(tx)).Create(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}

		// Create initial service state
		nextCheck := time.Now().Add(params.Interval.ToDuration())
		serviceState := &repo_service_states.CreateParams{
			ID:        utils.GenerateULID(),
			ServiceID: params.ID,
			Status:    models.StatusUnknown,
			NextCheck: &nextCheck,
		}

		_, err = s.store.ServiceStates(repos.WithTx(tx)).Create(ctx, *serviceState)
		if err != nil {
			return fmt.Errorf("failed to create initial service state: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create service in transaction: %w", err)
	}

	svcView, err := s.GetViewByID(ctx, params.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created service: %w", err)
	}

	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeCreated,
		svcView,
	))

	return svcView, nil
}

type UpdateParams = repo_services.UpdateServiceRequest

// Update service
func (s *Service) Update(ctx context.Context, id string, params UpdateParams) (*models.ServiceFullView, error) {
	item, err := s.store.Services().Update(ctx, id, params)
	if err != nil {
		return nil, err
	}

	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeUpdated,
		item,
	))

	return item, nil
}

// GetByID returns service by ID
func (s *Service) GetByID(ctx context.Context, id string) (*models.Service, error) {
	svc, err := s.store.Services().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storecmn.ErrNotFound
		}
		return nil, err
	}

	return svc, nil
}

// GetViewByID returns service view by ID
func (s *Service) GetViewByID(ctx context.Context, id string) (*models.ServiceFullView, error) {
	return s.store.Services().GetViewByID(ctx, id)
}

type FindParams = repo_services.FindParams

// FindView services by params
func (s *Service) FindView(ctx context.Context, params FindParams) (*dbutils.FindResponseWithCount[*models.ServiceFullView], error) {
	return s.store.Services().FindView(ctx, params)
}

// Delete service by ID
func (s *Service) Delete(ctx context.Context, id string) error {
	// Get service to find name for scheduler cleanup
	svc, err := s.store.Services().GetViewByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get service: %w", err)
	}

	// Delete service and related data in a transaction
	err = dbutils.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
		if err := s.store.ServiceStates(repos.WithTx(tx)).DeleteByServiceID(ctx, id); err != nil {
			return fmt.Errorf("failed to delete service states: %w", err)
		}

		if err := s.store.Incidents(repos.WithTx(tx)).DeleteByServiceID(ctx, id); err != nil {
			return fmt.Errorf("failed to delete incidents: %w", err)
		}

		if err := s.store.Services(repos.WithTx(tx)).Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete service: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to delete service in transaction: %w", err)
	}

	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeDeleted,
		svc,
	))

	return nil
}

// Exists checks if there are any services
func (s *Service) Exists(ctx context.Context, id string) (bool, error) {
	res, err := s.store.Services().Exist(ctx, id)
	if err != nil {
		return false, fmt.Errorf("failed to check if services exist: %w", err)
	}
	return res > 0, nil
}
