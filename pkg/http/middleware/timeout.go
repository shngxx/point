package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Timeout returns a middleware that sets a timeout for request processing
func Timeout(timeout time.Duration) Handler {
	return func(c *fiber.Ctx) error {
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()

		// Update context in Fiber
		c.SetUserContext(ctx)

		// Check if context is done
		done := make(chan error, 1)
		go func() {
			done <- c.Next()
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			return fiber.NewError(fiber.StatusRequestTimeout, "Request timeout")
		}
	}
}

