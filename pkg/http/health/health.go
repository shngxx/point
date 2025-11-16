package health

import (
	"github.com/gofiber/fiber/v2"
)

// LivenessHandler handles liveness probe requests
// Returns 200 OK if the server is alive
func LivenessHandler(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "alive",
	})
}

