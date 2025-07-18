package web

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

// handleGetAllTags handles GET /api/v1/tags
//
//	@Summary		Get all tags
//	@Description	Retrieves all unique tags used across services
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		string			"List of unique tags"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/tags [get]
func (h *Server) handleGetAllTags(c *fiber.Ctx) error {
	tags, err := h.storage.GetAllTags(c.Context())
	if err != nil {
		return fmt.Errorf("failed to get all tags: %w", err)
	}
	return c.JSON(tags)
}

// handleGetAllTagsWithCount handles GET /api/v1/tags/count
//
//	@Summary		Get all tags with usage count
//	@Description	Retrieves all unique tags along with their usage count across services
//	@Tags			tags
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]int	"Map of tags with their usage count"
//	@Failure		500	{object}	ErrorResponse	"Internal server error"
//	@Router			/tags/count [get]
func (h *Server) handleGetAllTagsWithCount(c *fiber.Ctx) error {
	tagsWithCount, err := h.storage.GetAllTagsWithCount(c.Context())
	if err != nil {
		return fmt.Errorf("failed to get tags with count: %w", err)
	}
	return c.JSON(tagsWithCount)
}
