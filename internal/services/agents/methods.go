package agents

import (
	"context"
	"fmt"
	"slices"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
)

type CreateParams struct {
	Name        string
	Description *string
	Host        string
	Port        int64
	TokenCt     []byte
	TokenNonce  []byte
	TokenHint   string
	Tags        []string
	Config      map[string]any
}

// Create
func (s *Service) Create(ctx context.Context, params CreateParams) (*models.Agent, error) {
	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	// Convert tags to JSONField
	tags := storecmn.JSONField("[]")
	if err := tags.UnmarshalAny(params.Tags); err != nil {
		return nil, fmt.Errorf("failed to convert tags to json raw message: %w", err)
	}

	// Convert config to JSONField
	config := storecmn.JSONField("{}")
	if err := config.UnmarshalAny(params.Config); err != nil {
		return nil, fmt.Errorf("failed to convert config to json raw message: %w", err)
	}

	createParams := repo_agents.CreateParams{
		ID:          utils.GenerateULID(),
		Name:        params.Name,
		Description: params.Description,
		Host:        params.Host,
		Port:        params.Port,
		TokenCt:     params.TokenCt,
		TokenNonce:  params.TokenNonce,
		TokenHint:   params.TokenHint,
		Tags:        tags,
		Config:      config,
	}

	return s.store.Agents().Create(ctx, createParams)
}

// Delete
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.store.Agents().Delete(ctx, id)
}

// GetByID
func (s *Service) GetByID(ctx context.Context, id string) (*models.Agent, error) {
	return s.store.Agents().GetByID(ctx, id)
}

// GetAll
func (s *Service) GetAll(ctx context.Context) ([]*models.Agent, error) {
	return s.store.Agents().GetAll(ctx)
}
