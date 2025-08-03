package web

import "github.com/gofiber/fiber/v2"

// ErrorResponse represents an error response
//
//	@Description	Error response
type ErrorResponse struct {
	Error string `json:"error" example:"Error description"`
}

// newErrorResponse creates a new ErrorResponse and sends it as a JSON response
func newErrorResponse(c *fiber.Ctx, status int, err error) error {
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
