package response

import "github.com/gofiber/fiber/v2"

// Writer provides helper methods for writing responses
type Writer struct {
	ctx *fiber.Ctx
}

// NewWriter creates a new response writer
func NewWriter(c *fiber.Ctx) *Writer {
	return &Writer{ctx: c}
}

// JSON sends a JSON response
func (w *Writer) JSON(status int, data any) error {
	return w.ctx.Status(status).JSON(data)
}

// String sends a plain text response
func (w *Writer) String(status int, s string) error {
	return w.ctx.Status(status).SendString(s)
}

// Send sends a response with the given body
func (w *Writer) Send(status int, body []byte) error {
	return w.ctx.Status(status).Send(body)
}
