package middleware

import (
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// Recovery returns a middleware that recovers from panics
// It logs the panic and returns a 500 Internal Server Error
func Recovery() Handler {
	recoverMiddleware := recover.New(recover.Config{
		EnableStackTrace: true,
	})
	return ToFiber(recoverMiddleware)
}
