package web

import (
	"github.com/gofiber/fiber/v2"
	"github.com/sxwebdev/sentinel/internal/services/notifications"
	"github.com/sxwebdev/sentinel/internal/store/storecmn"
)

// notificationProviderCreate creates a new notification provider
//
//	@Summary		Create new notification provider
//	@Description	Creates a new notification provider with the given configuration.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			service	body		notifications.CreateProviderParams	true	"Body params"
//	@Success		201		{object}	models.NotificationProvider			"Notification provider created"
//	@Failure		400		{object}	ErrorResponse						"Bad request"
//	@Failure		500		{object}	ErrorResponse						"Internal server error"
//	@Router			/settings/notifications/providers [post]
func (s *Server) notificationProviderCreate(c *fiber.Ctx) error {
	var data notifications.CreateProviderParams
	if err := c.BodyParser(&data); err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	// Add service
	svc, err := s.baseServices.Notifications().Providers().Create(c.Context(), data)
	if err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	return c.Status(fiber.StatusCreated).JSON(svc)
}

// notificationProviderUpdate updates a notification provider by ID
//
//	@Summary		Update a notification provider
//	@Description	Updates a notification provider with the given ID and configuration.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string								true	"Notification Provider ID"
//	@Param			service	body		notifications.UpdateProviderParams	true	"Body params"
//	@Success		200		{object}	models.NotificationProvider			"Notification provider updated"
//	@Failure		400		{object}	ErrorResponse						"Bad request"
//	@Failure		500		{object}	ErrorResponse						"Internal server error"
//	@Router			/settings/notifications/providers/{id} [put]
func (s *Server) notificationProviderUpdate(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return newErrorResponse(c, fiber.StatusBadRequest, storecmn.ErrEmptyID)
	}

	var data notifications.UpdateProviderParams
	if err := c.BodyParser(&data); err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	// Update service
	svc, err := s.baseServices.Notifications().Providers().Update(c.Context(), id, data)
	if err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	return c.JSON(svc)
}

// notificationProviderList lists all notification providers
//
//	@Summary		List all notification providers
//	@Description	Retrieves a list of all configured notification providers.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		models.NotificationProvider	"List of notification providers"
//	@Failure		500	{object}	ErrorResponse				"Internal server error"
//	@Router			/settings/notifications/providers [get]
func (s *Server) notificationProviderList(c *fiber.Ctx) error {
	providers, err := s.baseServices.Notifications().Providers().GetAll(c.Context())
	if err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return c.JSON(providers)
}

// notificationProviderDelete deletes a notification provider by ID
//
//	@Summary		Delete a notification provider
//	@Description	Deletes a notification provider by its ID.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Notification Provider ID"
//	@Success		204
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/settings/notifications/providers/{id} [delete]
func (s *Server) notificationProviderDelete(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return newErrorResponse(c, fiber.StatusBadRequest, storecmn.ErrEmptyID)
	}

	if err := s.baseServices.Notifications().Providers().Delete(c.Context(), id); err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// notificationProviderTest tests a notification provider by ID
//
//	@Summary		Test a notification provider
//	@Description	Sends a test notification using the specified provider ID.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			id	path	string	true	"Notification Provider ID"
//	@Success		200
//	@Failure		400	{object}	ErrorResponse	"Bad request"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/settings/notifications/providers/{id}/test [post]
func (s *Server) notificationProviderTest(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return newErrorResponse(c, fiber.StatusBadRequest, storecmn.ErrEmptyID)
	}

	if err := s.baseServices.Notifications().Providers().Test(c.Context(), id); err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return c.SendStatus(fiber.StatusOK)
}

// notificationHistoryList lists notification history with optional filters and pagination
//
//	@Summary		List notification history
//	@Description	Retrieves a list of notification history records with optional filtering by status and pagination.
//	@Tags			notifications
//	@Accept			json
//	@Produce		json
//	@Param			status		query		string															false	"Filter by status (e.g., 'sent', 'failed')"
//	@Param			order_by	query		string															false	"Order by field (default is 'created_at')"
//	@Param			page		query		int32															false	"Page number for pagination (default is 1)"
//	@Param			page_size	query		int32															false	"Number of items per page (default is 20)"
//	@Success		200			{object}	storecmn.FindResponseWithCount[models.NotificationHistoryView]	"List of notification history records"
//	@Failure		400			{object}	ErrorResponse													"Bad request"
//	@Failure		500			{object}	ErrorResponse													"Internal server error"
//	@Router			/settings/notifications/history [get]
func (s *Server) notificationHistoryList(c *fiber.Ctx) error {
	var params notifications.FindHistoryParams
	if err := c.QueryParser(&params); err != nil {
		return newErrorResponse(c, fiber.StatusBadRequest, err)
	}

	histories, err := s.baseServices.Notifications().History().Find(c.Context(), notifications.FindHistoryParams{
		Status:   params.Status,
		OrderBy:  params.OrderBy,
		Page:     params.Page,
		PageSize: params.PageSize,
	})
	if err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return c.JSON(histories)
}
