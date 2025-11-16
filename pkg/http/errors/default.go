package errors

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// DefaultErrorHandler is the default error handler implementation
type DefaultErrorHandler struct{}

// NewDefaultErrorHandler creates a new default error handler
func NewDefaultErrorHandler() ErrorHandler {
	return &DefaultErrorHandler{}
}

// Handle processes errors and returns appropriate HTTP responses
func (h *DefaultErrorHandler) Handle(c *fiber.Ctx, err error) error {
	// Check if it's a Fiber error
	if fiberErr, ok := err.(*fiber.Error); ok {
		return c.Status(fiberErr.Code).JSON(ErrorResponse{
			Success: false,
			Error:   fiberErr.Message,
			Code:    getErrorCode(fiberErr.Code),
		})
	}

	// Default to 500 Internal Server Error
	return c.Status(http.StatusInternalServerError).JSON(ErrorResponse{
		Success: false,
		Error:   err.Error(),
		Code:    CodeInternalError,
	})
}

// getErrorCode maps HTTP status codes to error codes
func getErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return CodeBadRequest
	case http.StatusUnauthorized:
		return CodeUnauthorized
	case http.StatusForbidden:
		return CodeForbidden
	case http.StatusNotFound:
		return CodeNotFound
	case http.StatusRequestTimeout:
		return CodeTimeout
	default:
		return CodeInternalError
	}
}

