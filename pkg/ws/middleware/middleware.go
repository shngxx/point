package middleware

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

// Handler is a middleware handler function
// It receives a Connection and returns an error
type Handler func(c ConnectionInterface) error

// Chain chains multiple middleware handlers together
func Chain(handlers ...Handler) Handler {
	return func(c ConnectionInterface) error {
		for _, handler := range handlers {
			if err := handler(c); err != nil {
				return err
			}
		}
		return nil
	}
}
