package http

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	httperrors "github.com/shngxx/point/pkg/http/errors"
	"github.com/shngxx/point/pkg/http/health"
	"github.com/shngxx/point/pkg/http/hooks"
	"github.com/shngxx/point/pkg/http/middleware"
	"github.com/shngxx/point/pkg/http/routing"
	"github.com/shngxx/point/pkg/http/shutdown"
	httpvalidation "github.com/shngxx/point/pkg/http/validation"
)

// Context wraps fiber.Ctx for convenience
type Context = fiber.Ctx

// Handler is a request handler function
type Handler func(c *Context) error

// ErrorHandler is an alias for the error handler interface
type ErrorHandler = httperrors.ErrorHandler

// Validator is an alias for the validator interface
type Validator = httpvalidation.Validator

// Server represents the HTTP server wrapper
type Server struct {
	app          *fiber.App
	address      string
	config       ServerConfig
	logger       *zerolog.Logger
	middleware   []middleware.Handler
	errorHandler ErrorHandler
	healthCheck  func() error
	validator    Validator
	hookManager  *hooks.Manager
}

// New creates a new Server instance with the given options
func New(opts ...Option) *Server {
	nop := zerolog.Nop()
	s := &Server{
		logger:       &nop,
		errorHandler: httperrors.NewDefaultErrorHandler(),
		config:       &DefaultConfig{},
		hookManager:  hooks.NewManager(),
	}

	// Apply options
	for _, opt := range opts {
		opt(s)
	}

	// Build address from config if not set
	if s.address == "" {
		if s.config.GetPort() > 0 {
			s.address = fmt.Sprintf(":%d", s.config.GetPort())
		} else {
			s.address = s.config.GetAddress()
		}
	}

	// Initialize Fiber app
	s.app = fiber.New(fiber.Config{
		ReadTimeout:  s.config.GetReadTimeout(),
		WriteTimeout: s.config.GetWriteTimeout(),
		IdleTimeout:  s.config.GetIdleTimeout(),
		ErrorHandler: s.errorHandler.Handle,
	})

	// Register global middleware
	for _, mw := range s.middleware {
		s.app.Use(middleware.ToFiber(mw))
	}

	// Register health check endpoints
	s.app.Get("/health", health.LivenessHandler)
	s.app.Get("/ready", health.ReadinessHandler(s.healthCheck))

	return s
}

// NewWithDefaults creates a new HTTP server with default middleware stack
// This is a convenience function that sets up Recovery, Logger, and RequestID middleware automatically
func NewWithDefaults(cfg ServerConfig, l *zerolog.Logger) *Server {
	return New(
		WithConfig(cfg),
		WithLogger(l),
		WithMiddleware(
			middleware.Recovery(),
			middleware.Logger(l),
			middleware.RequestID(),
		),
	)
}

// Use registers global middleware
func (s *Server) Use(mw ...middleware.Handler) {
	for _, m := range mw {
		s.app.Use(middleware.ToFiber(m))
	}
}

// GET registers a GET route
func (s *Server) GET(path string, handler Handler) {
	s.app.Get(path, handler)
}

// POST registers a POST route
func (s *Server) POST(path string, handler Handler) {
	s.app.Post(path, handler)
}

// PUT registers a PUT route
func (s *Server) PUT(path string, handler Handler) {
	s.app.Put(path, handler)
}

// DELETE registers a DELETE route
func (s *Server) DELETE(path string, handler Handler) {
	s.app.Delete(path, handler)
}

// PATCH registers a PATCH route
func (s *Server) PATCH(path string, handler Handler) {
	s.app.Patch(path, handler)
}

// Group creates a new route group
func (s *Server) Group(prefix string, fn func(*routing.Group)) {
	group := routing.NewGroup(s.app, prefix)
	fn(group)
}

// run starts the server and blocks until shutdown
func (s *Server) run() error {
	// Execute BeforeStart hooks
	if err := s.hookManager.Execute(hooks.BeforeStart); err != nil {
		return fmt.Errorf("before start hook failed: %w", err)
	}

	// Start server in a goroutine
	errChan := make(chan error, 1)
		go func() {
		s.logger.Info().Str("address", s.address).Msg("Starting server")
		if err := s.app.Listen(s.address); err != nil {
			errChan <- err
		}
	}()

	// Wait a bit to ensure server started
	time.Sleep(100 * time.Millisecond)

	// Execute AfterStart hooks
	if err := s.hookManager.Execute(hooks.AfterStart); err != nil {
		s.logger.Warn().Err(err).Msg("After start hook failed")
	}

	// Wait for shutdown signal
	sig := shutdown.WaitForShutdownSignal()
	s.logger.Info().Str("signal", sig.String()).Msg("Received shutdown signal")

	// Execute BeforeShutdown hooks
	if err := s.hookManager.Execute(hooks.BeforeShutdown); err != nil {
		s.logger.Warn().Err(err).Msg("Before shutdown hook failed")
	}

	// Graceful shutdown
	if err := shutdown.GracefulShutdown(s.app, s.config.GetShutdownTimeout()); err != nil {
		s.logger.Error().Err(err).Msg("Shutdown error")
		return err
	}

	// Execute AfterShutdown hooks
	if err := s.hookManager.Execute(hooks.AfterShutdown); err != nil {
		s.logger.Warn().Err(err).Msg("After shutdown hook failed")
	}

	s.logger.Info().Msg("Server stopped gracefully")

	return nil
}

// Start starts the server and exits the program if an error occurs
// This is a convenience method for applications that want to exit on server errors
// It logs the error using the server's logger before exiting
func (s *Server) Start() {
	if err := s.run(); err != nil {
		s.logger.Error().Err(err).Msg("Server startup error")
		os.Exit(1)
	}
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return shutdown.GracefulShutdown(s.app, s.config.GetShutdownTimeout())
}

// App returns the underlying Fiber app (for advanced use cases)
func (s *Server) App() *fiber.App {
	return s.app
}
