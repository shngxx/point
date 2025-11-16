package ws

import (
	"encoding/json"
	"sync"
)

// Message represents a generic WebSocket message
type Message struct {
	Action string          `json:"action"`
	Data   json.RawMessage `json:"data,omitempty"`
	Type   string          `json:"type,omitempty"`
}

// MessageHandler is a function that handles a message
type MessageHandler func(conn *Connection, message *Message) error

// Router handles message routing by action/type
type Router struct {
	handlers map[string]MessageHandler
	mu       sync.RWMutex
}

// NewRouter creates a new message router
func NewRouter() *Router {
	return &Router{
		handlers: make(map[string]MessageHandler),
	}
}

// Handle registers a handler for a specific action
func (r *Router) Handle(action string, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[action] = handler
}

// Route routes a message to the appropriate handler
func (r *Router) Route(conn *Connection, message *Message) error {
	r.mu.RLock()
	handler, ok := r.handlers[message.Action]
	r.mu.RUnlock()

	if !ok {
		// Try type field if action not found
		if message.Type != "" {
			r.mu.RLock()
			handler, ok = r.handlers[message.Type]
			r.mu.RUnlock()
		}
	}

	if !ok {
		return ErrUnknownAction
	}

	return handler(conn, message)
}

// HasHandler checks if a handler exists for the given action
func (r *Router) HasHandler(action string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.handlers[action]
	return ok
}

// Errors
var (
	ErrUnknownAction = &Error{Code: "UNKNOWN_ACTION", Message: "Unknown message action"}
)

// Error represents a WebSocket error
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *Error) Error() string {
	return e.Message
}

