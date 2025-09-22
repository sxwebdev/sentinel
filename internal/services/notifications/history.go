package notifications

import (
	"context"
	"fmt"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_notification_history"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/mx/logger"
)

type History struct {
	logger logger.Logger
	store  *store.Store
	sender *Sender
}

func newHistory(l logger.Logger, store *store.Store, sender *Sender) *History {
	return &History{
		logger: l,
		store:  store,
		sender: sender,
	}
}

// SendAlert sends an alert notification to all enabled providers
func (s *History) SendAlert(ctx context.Context, incidentID, message string) error {
	// Get all enabled providers
	providers, err := s.store.NotificationProviders().GetAllEnabled(ctx)
	if err != nil {
		return fmt.Errorf("failed to get enabled providers: %w", err)
	}

	// Send alert to each provider
	for _, provider := range providers {
		params := CreateHistoryParams{
			ProviderID: provider.ID,
			Message:    message,
		}

		if incidentID != "" {
			params.IncidentID = &incidentID
		}

		if _, err := s.Create(ctx, params); err != nil {
			return fmt.Errorf("failed to send alert to provider %s: %v", provider.ID, err)
		}
	}

	return nil
}

type CreateHistoryParams struct {
	ProviderID string  `db:"provider_id" json:"provider_id"`
	IncidentID *string `db:"incident_id" json:"incident_id"`
	Message    string  `db:"message" json:"message"`
}

// Validate validates the CreateHistoryParams fields
func (s CreateHistoryParams) Validate() error {
	if s.ProviderID == "" {
		return storecmn.ErrEmptyID
	}

	if s.Message == "" {
		return fmt.Errorf("empty message")
	}

	return nil
}

// Create creates a new notification history record
func (s *History) Create(ctx context.Context, params CreateHistoryParams) (*models.NotificationHistory, error) {
	if err := params.Validate(); err != nil {
		return nil, err
	}

	createParams := repo_notification_history.CreateParams{
		ID:         utils.GenerateULID(),
		ProviderID: params.ProviderID,
		IncidentID: params.IncidentID,
		Message:    params.Message,
	}

	item, err := s.store.NotificationHistory().Create(ctx, createParams)
	if err != nil {
		return nil, err
	}

	s.sender.looper.Trigger(ctx)

	return item, nil
}

// Delete deletes a notification history record by ID
func (s *History) Delete(ctx context.Context, id string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	return s.store.NotificationHistory().Delete(ctx, id)
}
