package agents

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
)

type CreateParams struct {
	Name        string
	Description *string
	TokenCt     []byte
	TokenNonce  []byte
	TokenHint   string
	Tags        []string
	Config      map[string]any
}

// Validate
func (p CreateParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}

	if p.TokenCt == nil {
		return fmt.Errorf("token_ct is required")
	}

	if p.TokenNonce == nil {
		return fmt.Errorf("token_nonce is required")
	}

	if p.TokenHint == "" {
		return fmt.Errorf("token_hint is required")
	}

	return nil
}

// Create a new agent
func (s *Service) Create(ctx context.Context, params CreateParams) (*models.Agent, error) {
	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
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

	createParams := repo_agents.CreateParams{
		ID:          utils.GenerateULID(),
		Name:        params.Name,
		Description: params.Description,
		TokenCt:     params.TokenCt,
		TokenNonce:  params.TokenNonce,
		TokenHint:   params.TokenHint,
		Tags:        tags,
		Config:      config,
	}

	return s.store.Agents().Create(ctx, createParams)
}

// Delete an existing agent
func (s *Service) Delete(ctx context.Context, id string) error {
	return s.store.Agents().Delete(ctx, id)
}

// GetByID retrieves an agent by its ID
func (s *Service) GetByID(ctx context.Context, id string) (*models.Agent, error) {
	return s.store.Agents().GetByID(ctx, id)
}

type FindParams = repo_agents.FindParams

// Find all agents with given filters and pagination
func (s *Service) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Agent], error) {
	return s.store.Agents().Find(ctx, params)
}

type UpdateParams struct {
	Name         string
	Description  *string
	Fingerprint  *string
	Status       string
	IsEnabled    bool
	Tags         []string
	Config       map[string]any
	SystemInfo   models.SystemInfo
	LastOnlineAt *time.Time
	FieldMask    dbutils.FieldMask[repo_agents.ColumnName]
}

// Validate
func (p UpdateParams) Validate() error {
	if p.FieldMask.Contains(repo_agents.ColumnNameAgentsName) && p.Name == "" {
		return fmt.Errorf("name is required")
	}

	if p.FieldMask.Contains(repo_agents.ColumnNameAgentsFingerprint) && (p.Fingerprint == nil || *p.Fingerprint == "") {
		return fmt.Errorf("fingerprint is required")
	}

	return nil
}

// Update an existing agent
func (s *Service) Update(ctx context.Context, id string, params UpdateParams) (*models.Agent, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
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

	// Convert system info to JSONField
	systemInfo := storecmn.JSONField("{}")
	if err := systemInfo.UnmarshalFromAny(params.SystemInfo); err != nil {
		return nil, fmt.Errorf("failed to convert system info to json raw message: %w", err)
	}

	updateParams := repo_agents.UpdateRequest{
		Agent: models.Agent{
			Name:         params.Name,
			Description:  params.Description,
			Fingerprint:  params.Fingerprint,
			Status:       params.Status,
			IsEnabled:    params.IsEnabled,
			Tags:         tags,
			Config:       config,
			SystemInfo:   systemInfo,
			LastOnlineAt: params.LastOnlineAt,
		},
		FieldMask: params.FieldMask,
	}

	return s.store.Agents().Update(ctx, id, updateParams)
}
