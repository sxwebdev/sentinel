package web

import (
	"github.com/gofiber/fiber/v2"
)

// handleAPIDashboardStats returns dashboard statistics
//
//	@Summary		Get dashboard statistics
//	@Description	Returns statistics for the dashboard
//	@Tags			dashboard
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	DashboardStats	"Dashboard statistics"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/dashboard/stats [get]
func (s *Server) handleAPIDashboardStats(c *fiber.Ctx) error {
	stats, err := s.getDashboardStats(c.Context())
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse{Error: err.Error()})
	}

	return c.JSON(stats)
}
