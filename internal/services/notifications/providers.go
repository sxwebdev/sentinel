package notifications

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_notification_providers"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
)

type Providers struct {
	store *store.Store

	history *History
}

func newProviders(store *store.Store, history *History) *Providers {
	return &Providers{
		store:   store,
		history: history,
	}
}

// GetByID retrieves a notification provider by its ID
func (s *Providers) GetByID(ctx context.Context, id string) (*models.NotificationProvider, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	item, err := s.store.NotificationProviders().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storecmn.ErrNotFound
		}
		return nil, err
	}

	return item, nil
}

// GetAll retrieves all service states
func (s *Providers) GetAll(ctx context.Context) ([]*models.NotificationProvider, error) {
	return s.store.NotificationProviders().GetAll(ctx)
}

type CreateProviderParams struct {
	ProviderType models.NotificationProviderType `json:"provider_type" example:"shoutrrr"`
	Config       map[string]any                  `json:"config" example:"{\"urls\": [\"slack://hooks.slack.com/services/...\"]}"`
}

// Validate
func (s CreateProviderParams) Validate() error {
	if err := s.ProviderType.Validate(); err != nil {
		return err
	}

	// Convert config to JSONField
	config := storecmn.JSONField("{}")
	if err := config.UnmarshalFromAny(s.Config); err != nil {
		return fmt.Errorf("failed to convert config to json raw message: %w", err)
	}

	if err := validateProviderConfig(s.ProviderType, config); err != nil {
		return err
	}

	return nil
}

// Create creates a new notification provider
func (s *Providers) Create(ctx context.Context, params CreateProviderParams) (*models.NotificationProvider, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	// Convert config to JSONField
	config := storecmn.JSONField("{}")
	if err := config.UnmarshalFromAny(params.Config); err != nil {
		return nil, fmt.Errorf("failed to convert config to json raw message: %w", err)
	}

	createParams := repo_notification_providers.CreateParams{
		ID:           utils.GenerateULID(),
		ProviderType: params.ProviderType,
		Config:       config,
	}

	return s.store.NotificationProviders().Create(ctx, createParams)
}

type UpdateProviderParams struct {
	ProviderType models.NotificationProviderType `db:"provider_type" json:"provider_type"`
	Config       storecmn.JSONField              `db:"config" json:"config"`
	IsEnabled    bool                            `db:"is_enabled" json:"is_enabled"`
}

// Validate
func (s UpdateProviderParams) Validate() error {
	if err := s.ProviderType.Validate(); err != nil {
		return err
	}

	if err := validateProviderConfig(s.ProviderType, s.Config); err != nil {
		return err
	}

	return nil
}

// Update updates a notification provider by ID
func (s *Providers) Update(ctx context.Context, id string, params UpdateProviderParams) (*models.NotificationProvider, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	if err := params.Validate(); err != nil {
		return nil, err
	}

	updateParams := repo_notification_providers.UpdateParams{
		ID:           id,
		ProviderType: params.ProviderType,
		Config:       params.Config,
		IsEnabled:    params.IsEnabled,
	}

	if err := s.store.NotificationProviders().Update(ctx, updateParams); err != nil {
		return nil, err
	}

	return s.GetByID(ctx, id)
}

// Delete deletes a notification provider by ID
func (s *Providers) Delete(ctx context.Context, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	return s.store.NotificationProviders().Delete(ctx, id)
}

// Test tests a notification provider by ID
func (s *Providers) Test(ctx context.Context, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	// Get provider
	provider, err := s.store.NotificationProviders().GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Create test notification history record
	_, err = s.history.Create(ctx, CreateHistoryParams{
		ProviderID: provider.ID,
		Message:    "Test notification sent",
	})

	return err
}
