package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/google/uuid"
)

// RequestID returns a middleware that generates and sets a request ID
// The request ID is available in the context and can be retrieved using GetRequestID
func RequestID() Handler {
	requestIDMiddleware := requestid.New(requestid.Config{
		Generator: func() string {
			return uuid.New().String()
		},
		ContextKey: "request_id",
	})
	return ToFiber(requestIDMiddleware)
}

// GetRequestID retrieves the request ID from the context
func GetRequestID(c *fiber.Ctx) string {
	return c.Locals("request_id").(string)
}

