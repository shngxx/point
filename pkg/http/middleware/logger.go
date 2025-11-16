package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/rs/zerolog"
)

// Logger returns a middleware that logs HTTP requests
func Logger(l *zerolog.Logger) Handler {
	if l == nil {
		// Return no-op middleware if logger is nil
		return func(c *fiber.Ctx) error {
			return c.Next()
		}
	}

	// Use Fiber's logger middleware with custom output
	loggerMiddleware := fiberlogger.New(fiberlogger.Config{
		Format:     "${time} ${status} - ${latency} ${method} ${path}\n",
		TimeFormat: time.RFC3339,
		Output:     &logWriter{logger: l},
	})

	return ToFiber(loggerMiddleware)
}

// logWriter implements io.Writer to redirect logs to zerolog
type logWriter struct {
	logger *zerolog.Logger
}

func (w *logWriter) Write(p []byte) (n int, err error) {
	// Remove trailing newline if present
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.logger.Info().Msg(msg)
	return len(p), nil
}
