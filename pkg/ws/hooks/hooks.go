package hooks

import (
	"context"
)

// ConnectionInterface defines the interface for a WebSocket connection
// This avoids import cycles by not importing the ws package directly
type ConnectionInterface interface {
	SetMetadata(key string, value any)
	GetMetadata(key string) (any, bool)
	Subscribe(roomID string)
	Unsubscribe(roomID string)
	GetSubscriptions() []string
	IsSubscribed(roomID string) bool
	WriteJSON(v any) error
	Context() context.Context
	Close() error
}

// HookType represents the type of lifecycle hook
type HookType string

const (
	// OnConnect is called after a connection is established
	OnConnect HookType = "on_connect"

	// OnDisconnect is called before a connection is closed
	OnDisconnect HookType = "on_disconnect"

	// OnMessage is called before a message is processed
	OnMessage HookType = "on_message"

	// OnError is called when an error occurs
	OnError HookType = "on_error"

	// OnJoinRoom is called when a connection joins a room
	OnJoinRoom HookType = "on_join_room"

	// OnLeaveRoom is called when a connection leaves a room
	OnLeaveRoom HookType = "on_leave_room"
)

// HookFunc is a function that can be registered as a lifecycle hook
// It receives the connection and optional context data
type HookFunc func(conn ConnectionInterface, data ...any) error

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
func (m *Manager) Execute(hookType HookType, conn ConnectionInterface, data ...any) error {
	hooks, ok := m.hooks[hookType]
	if !ok {
		return nil
	}

	for _, hook := range hooks {
		if err := hook(conn, data...); err != nil {
			return err
		}
	}

	return nil
}
