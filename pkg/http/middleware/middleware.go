package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// Handler is a middleware handler function
// It receives a Fiber context and returns an error
type Handler func(c *fiber.Ctx) error

// ToFiber converts a middleware Handler to Fiber middleware
func ToFiber(h Handler) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return h(c)
	}
}

// Chain chains multiple middleware handlers together
func Chain(handlers ...Handler) Handler {
	return func(c *fiber.Ctx) error {
		for _, handler := range handlers {
			if err := handler(c); err != nil {
				return err
			}
		}
		return c.Next()
	}
}
