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
	"github.com/sxwebdev/sentinel/internal/store/repos"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/sxwebdev/sentinel/internal/utils"
	"github.com/tkcrm/modules/pkg/db/dbutils"
)

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

// GetByIDAndProjectID retrieves an agent by its ID
func (s *Service) GetByIDAndProjectID(ctx context.Context, id string, projectID string) (*models.Agent, error) {
	if id == "" {
		return nil, storecmn.ErrEmptyID
	}

	agent, err := s.store.Agents().GetByIDAndProjectID(ctx, id, projectID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, storecmn.ErrNotFound
		}
		return nil, err
	}

	return agent, nil
}

type CreateParams struct {
	Name        string
	Kind        models.AgentKindType
	Description string
	Tags        []string
	Config      models.AgentConfig
	IsEnabled   bool
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
func (s *Service) Create(ctx context.Context, params CreateParams, opts ...repos.Option) (*CreateResponse, error) {
	if len(params.Tags) > 0 {
		slices.Sort(params.Tags)
	}

	if err := params.Validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
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
		Tags:        params.Tags,
		IsEnabled:   params.IsEnabled,
		Config:      params.Config,
		ProjectID:   params.ProjectID,
	}

	item, err := s.store.Agents(opts...).Create(ctx, createParams)
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
	if id == "" {
		return storecmn.ErrEmptyID
	}

	// get agent
	agent, err := s.store.Agents().GetByIDAndProjectID(ctx, id, projectID)
	if err != nil {
		return err
	}

	// prevent deleting hub agents
	if agent.Kind == models.AgentKindTypeHub {
		return fmt.Errorf("hub agents cannot be deleted")
	}

	if err := s.store.Agents().Delete(ctx, id, projectID); err != nil {
		return err
	}

	s.dispatcher.Agents().Publish(dispatcher.NewBaseMessage(dispatcher.EventTypeDelete, projectID))

	return nil
}

type FindParams = repo_agents.FindParams

// Find all agents with given filters and pagination
func (s *Service) Find(ctx context.Context, params FindParams) (*storecmn.FindResponseWithCount[*models.Agent], error) {
	agents, err := s.store.Agents().Find(ctx, params)
	if err != nil {
		return nil, err
	}

	for _, agent := range agents.Items {
		if agent.Kind == models.AgentKindTypeHub {
			agent.TokenHint = "-"
			agent.Status = models.AgentStatusTypeActive
			agent.LastSeenAt = utils.Pointer(time.Now())
			agent.SystemInfo = *s.systemInfo
		}
	}

	return agents, nil
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

	updateParams := repo_agents.UpdateRequest{
		Agent: models.Agent{
			Name:        params.Name,
			Description: params.Description,
			Fingerprint: params.Fingerprint,
			Status:      params.Status,
			IsEnabled:   params.IsEnabled,
			Tags:        params.Tags,
			Config:      params.Config,
			SystemInfo:  params.SystemInfo,
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
func (s *Service) CheckAndUpsertFingerprint(ctx context.Context, id, projectID, fingerprint string) error {
	if id == "" {
		return storecmn.ErrEmptyID
	}

	if fingerprint == "" {
		return fmt.Errorf("fingerprint is required")
	}

	// Check if the fingerprint is already used by another agent
	existingAgent, err := s.GetByIDAndProjectID(ctx, id, projectID)
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
