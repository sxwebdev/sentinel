package web

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

// manualUpgradeHandler handles manual upgrade requests
//
//	@Summary		Manual upgrade
//	@Description	Triggers a manual upgrade of the server
//	@Tags			server
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	SuccessResponse	"Upgrade initiated successfully"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/server/upgrade [get]
func (s *Server) handleManualUpgrade(c *fiber.Ctx) error {
	if s.upgrader == nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, errors.New("upgrader service is not configured"))
	}

	if err := s.upgrader.Do(); err != nil {
		return newErrorResponse(c, fiber.StatusInternalServerError, err)
	}

	return newSuccessResponse(c, "Upgrade initiated successfully. Please check the logs for details.")
}

type healthCheckResponse struct {
	Status string `json:"status" example:"healthy"`
}

// healthCheckHandler handles health check requests
//
//	@Summary		Health check
//	@Description	Checks the health of the server
//	@Tags			server
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	healthCheckResponse	"Health check successful"
//	@Failure		500	{object}	ErrorResponse		"Internal server error"
//	@Router			/server/health [get]
func (s *Server) handleHealthCheck(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(healthCheckResponse{
		Status: "healthy",
	})
}
