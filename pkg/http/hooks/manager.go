package hooks

import (
	"fmt"
)

// Manager manages lifecycle hooks
type Manager struct {
	hooks map[HookType][]HookFunc
}

// NewManager creates a new hook manager
func NewManager() *Manager {
	return &Manager{
		hooks: make(map[HookType][]HookFunc),
	}
}

// Add registers a hook function for the given hook type
func (m *Manager) Add(hookType HookType, fn HookFunc) {
	if m.hooks == nil {
		m.hooks = make(map[HookType][]HookFunc)
	}
	m.hooks[hookType] = append(m.hooks[hookType], fn)
}

// Execute runs all hooks of the given type in order
// Returns the first error encountered, if any
func (m *Manager) Execute(hookType HookType) error {
	hooks, ok := m.hooks[hookType]
	if !ok {
		return nil
	}

	for _, hook := range hooks {
		if err := hook(); err != nil {
			return fmt.Errorf("hook %s failed: %w", hookType, err)
		}
	}

	return nil
}

