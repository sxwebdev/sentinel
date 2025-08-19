package web

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/sxwebdev/sentinel/internal/storage"
)

// ErrorResponse represents an error response
//
//	@Description	Error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error description"`
}

// newErrorResponse creates a new ErrorResponse and sends it as a JSON response
func newErrorResponse(c *fiber.Ctx, status int, err error) error {
	if errors.Is(err, storage.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	if errors.Is(err, storage.ErrAlreadyExists) {
		return c.Status(fiber.StatusConflict).JSON(ErrorResponse{
			Error: err.Error(),
		})
	}

	return c.Status(status).JSON(ErrorResponse{
		Error: err.Error(),
	})
}

// SuccessResponse represents a successful response
//
//	@Description	Successful response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// newSuccessResponse creates a new SuccessResponse and sends it as a JSON response
func newSuccessResponse(c *fiber.Ctx, message string) error {
	return c.Status(fiber.StatusOK).JSON(SuccessResponse{
		Message: message,
	})
}
