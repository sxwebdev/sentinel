package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sxwebdev/sentinel/internal/models"
	"github.com/sxwebdev/sentinel/internal/services/agents"
	"github.com/sxwebdev/sentinel/internal/store/repos/repo_agents"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
	"github.com/tkcrm/modules/pkg/db/dbutils"
)

type AgentCreateParams struct {
	Name        string
	Description *string
	Tags        []string
	Config      map[string]any
}

// agentsCreate creates a new agent
//
//	@Summary		Create new agent
//	@Description	Creates a new agent with the given configuration.
//	@Tags			agents
//	@Accept			json
//	@Produce		json
//	@Param			service	body		AgentCreateParams		true	"Body params"
//	@Success		201		{object}	agents.CreateResponse	"Agent created"
//	@Failure		400		{object}	ErrorResponse			"Bad request"
//	@Failure		500		{object}	ErrorResponse			"Internal server error"
//	@Router			/settings/agents [post]
func (s *Server) agentsCreate(c *fiber.Ctx) error {
	var data AgentCreateParams
	if err := c.BodyParser(&data); err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	params := agents.CreateParams{
		Name:        data.Name,
		Description: data.Description,
		Tags:        data.Tags,
		Config:      data.Config,
	}

	// Add service
	item, err := s.baseServices.Agents().Create(c.Context(), params)
	if err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(item)
}

// aggentsList lists all agents
//
//	@Summary		List agents
//	@Description	Retrieves a list of all agents.
//	@Tags			agents
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	storecmn.FindResponseWithCount[AgentDTO]	"List of agents"
//	@Failure		500	{object}	ErrorResponse								"Internal server error"
//	@Router			/settings/agents [get]
func (s *Server) agentsList(c *fiber.Ctx) error {
	data, err := s.baseServices.Agents().Find(c.Context(), agents.FindParams{})
	if err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	result := storecmn.FindResponseWithCount[AgentDTO]{
		Items: make([]AgentDTO, 0, len(data.Items)),
		Count: data.Count,
	}

	for _, item := range data.Items {
		dto, err := toAgentDTO(item)
		if err != nil {
			return newErrorResponse(c, fiber.StatusInternalServerError, err)
		}
		result.Items = append(result.Items, dto)
	}

	return c.JSON(result)
}

// agentDelete deletes an agent by ID
//
//	@Summary		Delete agent
//	@Description	Deletes an agent by its ID.
//	@Tags			agents
//	@Param			id	path	string	true	"Agent ID"
//	@Success		204
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		404	{object}	ErrorResponse	"Agent not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/settings/agents/{id} [delete]
func (s *Server) agentDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return newErrorResponse(c, fiber.StatusBadRequest, fiber.NewError(fiber.StatusBadRequest, "missing agent ID"))
	}

	if err := s.baseServices.Agents().Delete(c.Context(), id); err != nil {
		if err == storecmn.ErrNotFound {
			return newErrorResponse(c, fiber.StatusNotFound, err)
		}
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// agentGetByID retrieves an agent by ID
//
//	@Summary		Get agent by ID
//	@Description	Retrieves an agent by its ID.
//	@Tags			agents
//	@Param			id	path		string			true	"Agent ID"
//	@Success		200	{object}	AgentDTO		"Agent found"
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		404	{object}	ErrorResponse	"Agent not found"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/settings/agents/{id} [get]
func (s *Server) agentGetByID(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return newErrorResponse(c, fiber.StatusBadRequest, fiber.NewError(fiber.StatusBadRequest, "missing agent ID"))
	}

	svc, err := s.baseServices.Agents().GetByID(c.Context(), id)
	if err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	svcDTO, err := toAgentDTO(svc)
	if err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return c.JSON(svcDTO)
}

type AgentUpdateParams struct {
	Name        string
	Description *string
	IsEnabled   bool
	Tags        []string
	Config      map[string]any
}

// agentUpdate updates an existing agent
//
//	@Summary		Update agent
//	@Description	Updates an agent by its ID.
//	@Tags			agents
//	@Param			id		path		string				true	"Agent ID"
//	@Param			body	body		AgentUpdateParams	true	"Agent data"
//	@Success		200		{object}	AgentDTO			"Agent updated"
//	@Failure		400		{object}	ErrorResponse		"Bad request"
//	@Failure		404		{object}	ErrorResponse		"Agent not found"
//	@Failure		500		{object}	ErrorResponse		"Internal server error"
//	@Router			/settings/agents/{id} [put]
func (s *Server) agentUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return newErrorResponse(c, fiber.StatusBadRequest, fiber.NewError(fiber.StatusBadRequest, "missing agent ID"))
	}

	var req AgentUpdateParams
	if err := c.BodyParser(&req); err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	params := agents.UpdateParams{
		Name:        req.Name,
		Description: req.Description,
		IsEnabled:   req.IsEnabled,
		Tags:        req.Tags,
		Config:      req.Config,
		FieldMask: dbutils.FieldMask[repo_agents.ColumnName]{
			repo_agents.ColumnNameAgentsName,
			repo_agents.ColumnNameAgentsDescription,
			repo_agents.ColumnNameAgentsIsEnabled,
			repo_agents.ColumnNameAgentsTags,
			repo_agents.ColumnNameAgentsConfig,
		},
	}

	svc, err := s.baseServices.Agents().Update(c.Context(), id, params)
	if err != nil {
		if err == storecmn.ErrNotFound {
			return newErrorResponse(c, fiber.StatusNotFound, err)
		}
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	svcDTO, err := toAgentDTO(svc)
	if err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return c.JSON(svcDTO)
}

// toAgentDTO converts a models.Agent to AgentDTO
func toAgentDTO(m *models.Agent) (AgentDTO, error) {
	var tags []string
	if err := m.Tags.ConvertToAny(&tags); err != nil {
		return AgentDTO{}, err
	}

	var config map[string]any
	if err := m.Config.ConvertToAny(&config); err != nil {
		return AgentDTO{}, err
	}

	var systemInfo models.SystemInfo
	if err := m.SystemInfo.ConvertToAny(&systemInfo); err != nil {
		return AgentDTO{}, err
	}

	return AgentDTO{
		ID:           m.ID,
		Name:         m.Name,
		Description:  m.Description,
		TokenHint:    m.TokenHint,
		Fingerprint:  m.Fingerprint,
		Status:       m.Status,
		IsEnabled:    m.IsEnabled,
		Tags:         tags,
		Config:       config,
		SystemInfo:   systemInfo,
		LastOnlineAt: m.LastOnlineAt,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}, nil
}
