package agents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/sxwebdev/sentinel/internal/dispatcher"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
)

type CreateParams struct {
	Name        string
	Kind        models.AgentKindType
	Description string
	Tags        []string
	Config      models.AgentConfig
	ProjectID   string
}

// Validate
func (p CreateParams) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}

	if err := p.Kind.Validate(); err != nil {
		return fmt.Errorf("kind is invalid: %w", err)
	}

	if p.ProjectID == "" {
		return fmt.Errorf("project_id is required")
	}

	return nil
}

type CreateResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Token       string                 `json:"token"`
	TokenHint   string                 `json:"token_hint"`
	Status      models.AgentStatusType `json:"status"`
	IsEnabled   bool                   `json:"is_enabled"`
	Tags        []string               `json:"tags"`
	Config      models.AgentConfig     `json:"config"`
	CreatedAt   time.Time              `json:"created_at"`
}

// Create a new agent
func (s *Service) Create(ctx context.Context, params CreateParams) (*CreateResponse, error) {
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

	id := utils.GenerateULID()
	token, secret, tokenHint, err := NewAgentToken(id)
	if err != nil {
		return nil, fmt.Errorf("failed to generate agent token: %w", err)
	}

	// Hash the secret using Argon2id
	secretHash, err := HashSecretArgon2id(secret, DefaultArgon2)
	if err != nil {
		return nil, fmt.Errorf("failed to hash agent secret: %w", err)
	}

	createParams := repo_agents.CreateParams{
		ID:          id,
		Name:        params.Name,
		Description: params.Description,
		SecretHash:  secretHash,
		TokenHint:   tokenHint,
		Kind:        params.Kind,
		Tags:        tags,
		Config:      config,
		ProjectID:   params.ProjectID,
	}

	item, err := s.store.Agents().Create(ctx, createParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	result := CreateResponse{
		ID:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Token:       token,
		TokenHint:   tokenHint,
		Status:      item.Status,
		IsEnabled:   item.IsEnabled,
		Tags:        params.Tags,
		Config:      params.Config,
		CreatedAt:   item.CreatedAt,
	}

	s.dispatcher.Agents().Publish(dispatcher.NewBaseMessage(dispatcher.EventTypeCreate, params.ProjectID))

	return &result, nil
}

// Delete an existing agent
func (s *Service) Delete(ctx context.Context, id, projectID string) error {
	if err := s.store.Agents().Delete(ctx, id); err != nil {
		return err
	}

	s.dispatcher.Agents().Publish(dispatcher.NewBaseMessage(dispatcher.EventTypeDelete, projectID))

	return nil
}

// GetByID retrieves an agent by its ID
func (s *Service) GetByID(ctx context.Context, id string) (*models.Agent, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	agent, err := s.store.Agents().GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storecmn.ErrNotFound
		}
		return nil, err
	}

	return agent, nil
}

type FindParams = repo_agents.FindParams

// Find all agents with given filters and pagination
func (s *Service) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Agent], error) {
	return s.store.Agents().Find(ctx, params)
}

type UpdateParams struct {
	Name        string
	Description string
	Fingerprint *string
	Status      models.AgentStatusType
	IsEnabled   bool
	Tags        []string
	Config      models.AgentConfig
	SystemInfo  models.SystemInfo
	LastSeenAt  *time.Time
	FieldMask   dbutils.FieldMask[repo_agents.ColumnName]
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
func (s *Service) Update(ctx context.Context, id, projectID string, params UpdateParams) (*models.Agent, error) {
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
			Name:        params.Name,
			Description: params.Description,
			Fingerprint: params.Fingerprint,
			Status:      params.Status,
			IsEnabled:   params.IsEnabled,
			Tags:        tags,
			Config:      config,
			SystemInfo:  systemInfo,
			LastSeenAt:  params.LastSeenAt,
		},
		FieldMask: params.FieldMask,
	}

	item, err := s.store.Agents().Update(ctx, id, updateParams)
	if err != nil {
		return nil, err
	}

	s.dispatcher.Agents().Publish(dispatcher.NewBaseMessage(dispatcher.EventTypeUpdate, projectID))

	return item, nil
}

// CheckAndUpsertFingerprint checks if the fingerprint is unique and updates it if so
func (s *Service) CheckAndUpsertFingerprint(ctx context.Context, id, fingerprint string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	if fingerprint == "" {
		return fmt.Errorf("fingerprint is required")
	}

	// Check if the fingerprint is already used by another agent
	existingAgent, err := s.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if existingAgent.Fingerprint != nil && *existingAgent.Fingerprint != fingerprint {
		return fmt.Errorf("agent already has a different fingerprint")
	}

	// Update the agent with the new fingerprint
	updateParams := repo_agents.UpdateRequest{
		Agent: models.Agent{
			Fingerprint: &fingerprint,
			LastSeenAt:  utils.Pointer(time.Now()),
		},
		FieldMask: dbutils.FieldMask[repo_agents.ColumnName]{
			repo_agents.ColumnNameAgentsFingerprint,
			repo_agents.ColumnNameAgentsLastSeenAt,
		},
	}

	_, err = s.store.Agents().Update(ctx, id, updateParams)
	if err != nil {
		return err
	}

	s.dispatcher.Agents().Publish(
		dispatcher.NewBaseMessage(dispatcher.EventTypeUpdate, existingAgent.ProjectID),
	)

	return nil
}
