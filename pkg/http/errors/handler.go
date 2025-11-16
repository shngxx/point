package errors

import "github.com/gofiber/fiber/v2"

// ErrorHandler defines the interface for handling errors
type ErrorHandler interface {
	// Handle processes an error and returns a response
	Handle(c *fiber.Ctx, err error) error
}

