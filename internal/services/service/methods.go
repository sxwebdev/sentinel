package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"

	"github.com/go-playground/validator/v10"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/receiver"
	"github.com/sxwebdev/sentinel/internal/store/repos"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_service_states"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_services"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
)

type CreateUpdateParams struct {
	Name      string                     `validate:"required"`
	Protocol  models.ServiceProtocolType `validate:"required"`
	Interval  int64                      `validate:"required"`
	Timeout   int64                      `validate:"required"`
	Retries   int64                      `validate:"required,gte=0"`
	Tags      []string
	Config    map[string]any
	IsEnabled bool
}

// Create new service
func (s *Service) Create(ctx context.Context, params CreateUpdateParams) (*models.ServiceFullView, error) {
	if err := validator.New().Struct(params); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if err := params.Protocol.Validate(); err != nil {
		return nil, fmt.Errorf("invalid protocol: %w", err)
	}

	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	// Convert tags to JSONField
	tags := storecmn.JSONField("[]")
	if err := tags.UnmarshalFromAny(params.Tags); err != nil {
		return nil, fmt.Errorf("failed to convert tags to json raw message: %w", err)
	}

	// Convert config to JSONField
	config := storecmn.JSONField("{}")
	if err := config.UnmarshalFromAny(params.Config); err != nil {
		return nil, fmt.Errorf("failed to convert config to json raw message: %w", err)
	}

	createParams := repo_services.CreateParams{
		ID:        utils.GenerateULID(),
		Name:      params.Name,
		Protocol:  params.Protocol,
		Interval:  params.Interval,
		Timeout:   params.Timeout,
		Retries:   params.Retries,
		Tags:      tags,
		Config:    config,
		IsEnabled: params.IsEnabled,
	}

	err := storecmn.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
		// Create service
		_, err := s.store.Services(repos.WithTx(tx)).Create(ctx, createParams)
		if err != nil {
			return fmt.Errorf("failed to create service: %w", err)
		}

		// Create initial service state
		serviceState := &repo_service_states.CreateParams{
			ID:        utils.GenerateULID(),
			ServiceID: createParams.ID,
			Status:    models.ServiceStatusUnknown,
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
	if err := validator.New().Struct(params); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	if err := params.Protocol.Validate(); err != nil {
		return nil, fmt.Errorf("invalid protocol: %w", err)
	}

	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	// Convert tags to JSONField
	tags := storecmn.JSONField("[]")
	if err := tags.UnmarshalFromAny(params.Tags); err != nil {
		return nil, fmt.Errorf("failed to convert tags to json raw message: %w", err)
	}

	// Convert config to JSONField
	config := storecmn.JSONField("{}")
	if err := config.UnmarshalFromAny(params.Config); err != nil {
		return nil, fmt.Errorf("failed to convert config to json raw message: %w", err)
	}

	updateParams := repo_services.UpdateServiceRequest{
		Name:      params.Name,
		Protocol:  params.Protocol,
		Interval:  params.Interval,
		Timeout:   params.Timeout,
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
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	svc, err := s.store.Services().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storecmn.ErrNotFound
		}
		return nil, err
	}

	return svc, nil
}

// GetAllEnabled returns all enabled services
func (s *Service) GetAllEnabled(ctx context.Context) ([]*models.Service, error) {
	return s.store.Services().GetAllEnabled(ctx)
}

// GetAllEnabledByAgentID returns all enabled services assigned to the given agent ID
func (s *Service) GetAllEnabledByAgentID(ctx context.Context, agentID string) ([]*models.Service, error) {
	return s.store.Services().GetAllEnabledByAgentID(ctx, agentID)
}

// GetAllEnabledWithoutAgents returns all enabled services that are not assigned to any agents
func (s *Service) GetAllEnabledWithoutAgents(ctx context.Context) ([]*models.Service, error) {
	return s.store.Services().GetAllEnabledWithoutAgents(ctx)
}

// GetViewByID returns service view by ID
func (s *Service) GetViewByID(ctx context.Context, id string) (*models.ServiceFullView, error) {
	return s.store.Services().GetViewByID(ctx, id)
}

type FindParams = repo_services.FindParams

// FindView services by params
func (s *Service) FindView(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.ServiceFullView], error) {
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
	err = storecmn.WrapTx(ctx, s.store.SQLite(), func(tx *sql.Tx) error {
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
