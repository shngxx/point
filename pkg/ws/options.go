package ws

import (
	"github.com/rs/zerolog"
	"github.com/shngxx/point/pkg/ws/hooks"
	"github.com/shngxx/point/pkg/ws/middleware"
)

// Option is a function that configures the Manager
type Option func(*Manager)

// WithLogger sets a custom logger
func WithLogger(l *zerolog.Logger) Option {
	return func(m *Manager) {
		if l != nil {
			m.logger = l
		}
	}
}

// WithConfig sets a custom manager configuration
func WithConfig(cfg ManagerConfig) Option {
	return func(m *Manager) {
		if cfg != nil {
			m.config = cfg
		}
	}
}

// WithMiddleware sets global middleware
func WithMiddleware(mw ...middleware.Handler) Option {
	return func(m *Manager) {
		m.middleware = append(m.middleware, mw...)
	}
}

// WithHook registers a lifecycle hook
func WithHook(hookType hooks.HookType, fn hooks.HookFunc) Option {
	return func(m *Manager) {
		if m.hookManager == nil {
			m.hookManager = hooks.NewManager()
		}
		m.hookManager.Add(hookType, fn)
	}
}
