package http

import (
	"github.com/shngxx/point/pkg/http/hooks"
)

// AddHook registers a lifecycle hook
func (s *Server) AddHook(hookType hooks.HookType, fn hooks.HookFunc) {
	if s.hookManager == nil {
		s.hookManager = hooks.NewManager()
	}
	s.hookManager.Add(hookType, fn)
}

