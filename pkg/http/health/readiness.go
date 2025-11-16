package health

import (
	"github.com/gofiber/fiber/v2"
)

// ReadinessHandler handles readiness probe requests
// Returns 200 OK if the server is ready to accept traffic
// Returns 503 Service Unavailable if not ready
func ReadinessHandler(check func() error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if check == nil {
			// No check function means always ready
			return c.JSON(fiber.Map{
				"status": "ready",
			})
		}

		if err := check(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not ready",
				"error":  err.Error(),
			})
		}

		return c.JSON(fiber.Map{
			"status": "ready",
		})
	}
}

