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

type CreateUpdateParams struct {
	Name      string
	Protocol  models.ServiceProtocolType
	Interval  time.Duration
	Timeout   time.Duration
	Retries   int64
	Tags      []string
	Config    map[string]any
	IsEnabled bool
}

// Create new service
func (s *Service) Create(ctx context.Context, params CreateUpdateParams) (*models.ServiceFullView, error) {
	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	// Convert tags to JSONField
	tags := dbutils.JSONField("[]")
	if err := tags.UnmarshalAny(params.Tags); err != nil {
		return nil, fmt.Errorf("failed to convert tags to json raw message: %w", err)
	}

	// Convert config to JSONField
	config := dbutils.JSONField("{}")
	if err := config.UnmarshalAny(params.Config); err != nil {
		return nil, fmt.Errorf("failed to convert config to json raw message: %w", err)
	}

	createParams := repo_services.CreateParams{
		ID:        utils.GenerateULID(),
		Name:      params.Name,
		Protocol:  params.Protocol,
		Interval:  dbutils.Duration(params.Interval),
		Timeout:   dbutils.Duration(params.Timeout),
		Retries:   params.Retries,
		Tags:      tags,
		Config:    config,
		IsEnabled: params.IsEnabled,
	}

	err := dbutils.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
		// Create service
		_, err := s.store.Services(repos.WithTx(tx)).Create(ctx, createParams)
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}

		// Create initial service state
		nextCheck := time.Now().Add(params.Interval)
		serviceState := &repo_service_states.CreateParams{
			ID:        utils.GenerateULID(),
			ServiceID: createParams.ID,
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

	svcView, err := s.GetViewByID(ctx, createParams.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get created service: %w", err)
	}

	s.receiver.TriggerService().Publish(*receiver.NewTriggerServiceData(
		receiver.TriggerServiceEventTypeCreated,
		svcView,
	))

	return svcView, nil
}

// Update service
func (s *Service) Update(ctx context.Context, id string, params CreateUpdateParams) (*models.ServiceFullView, error) {
	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	// Convert tags to JSONField
	tags := dbutils.JSONField("[]")
	if err := tags.UnmarshalAny(params.Tags); err != nil {
		return nil, fmt.Errorf("failed to convert tags to json raw message: %w", err)
	}

	// Convert config to JSONField
	config := dbutils.JSONField("{}")
	if err := config.UnmarshalAny(params.Config); err != nil {
		return nil, fmt.Errorf("failed to convert config to json raw message: %w", err)
	}

	updateParams := repo_services.UpdateServiceRequest{
		Name:      params.Name,
		Protocol:  params.Protocol,
		Interval:  dbutils.Duration(params.Interval),
		Timeout:   dbutils.Duration(params.Timeout),
		Retries:   params.Retries,
		Tags:      tags,
		Config:    config,
		IsEnabled: params.IsEnabled,
	}

	item, err := s.store.Services().Update(ctx, id, updateParams)
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
