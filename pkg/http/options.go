package http

import (
	"github.com/rs/zerolog"
	"github.com/shngxx/point/pkg/http/middleware"
)

// Option is a function that configures the Server
type Option func(*Server)

// WithAddress sets the server address
func WithAddress(addr string) Option {
	return func(s *Server) {
		s.address = addr
	}
}

// WithLogger sets a custom logger
func WithLogger(l *zerolog.Logger) Option {
	return func(s *Server) {
		if l != nil {
			s.logger = l
		}
	}
}

// WithConfig sets a custom server configuration
func WithConfig(cfg ServerConfig) Option {
	return func(s *Server) {
		if cfg != nil {
			s.config = cfg
		}
	}
}

// WithMiddleware sets global middleware
func WithMiddleware(mw ...middleware.Handler) Option {
	return func(s *Server) {
		s.middleware = append(s.middleware, mw...)
	}
}

// WithErrorHandler sets a custom error handler
func WithErrorHandler(handler ErrorHandler) Option {
	return func(s *Server) {
		if handler != nil {
			s.errorHandler = handler
		}
	}
}

// WithHealthCheck sets a custom health check function for readiness probe
func WithHealthCheck(check func() error) Option {
	return func(s *Server) {
		s.healthCheck = check
	}
}

// WithValidator sets a custom validator
func WithValidator(validator Validator) Option {
	return func(s *Server) {
		if validator != nil {
			s.validator = validator
		}
	}
}
