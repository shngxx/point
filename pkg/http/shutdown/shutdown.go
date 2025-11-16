package shutdown

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// GracefulShutdown gracefully shuts down the Fiber app with a timeout
func GracefulShutdown(app *fiber.App, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return app.ShutdownWithContext(ctx)
}

