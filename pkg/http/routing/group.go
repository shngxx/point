package routing

import (
	"github.com/gofiber/fiber/v2"
	"github.com/shngxx/point/pkg/http/middleware"
)

// Handler is a request handler function
type Handler func(c *fiber.Ctx) error

// Group represents a route group
type Group struct {
	app    *fiber.App
	group  *fiber.Group
	prefix string
}

// NewGroup creates a new route group
func NewGroup(app *fiber.App, prefix string) *Group {
	group := app.Group(prefix)
	return &Group{
		app:    app,
		group:  group.(*fiber.Group),
		prefix: prefix,
	}
}

// Use registers middleware for this group
func (g *Group) Use(mw ...middleware.Handler) {
	for _, m := range mw {
		g.group.Use(middleware.ToFiber(m))
	}
}

// GET registers a GET route in this group
func (g *Group) GET(path string, handler Handler) {
	g.group.Get(path, handler)
}

// POST registers a POST route in this group
func (g *Group) POST(path string, handler Handler) {
	g.group.Post(path, handler)
}

// PUT registers a PUT route in this group
func (g *Group) PUT(path string, handler Handler) {
	g.group.Put(path, handler)
}

// DELETE registers a DELETE route in this group
func (g *Group) DELETE(path string, handler Handler) {
	g.group.Delete(path, handler)
}

// PATCH registers a PATCH route in this group
func (g *Group) PATCH(path string, handler Handler) {
	g.group.Patch(path, handler)
}

// Group creates a nested route group
func (g *Group) Group(prefix string, fn func(*Group)) {
	nested := NewGroup(g.app, g.prefix+prefix)
	fn(nested)
}
