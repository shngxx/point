package ws

import (
	"sync"

	"github.com/rs/zerolog"
)

// Room represents a named group of connections
type Room struct {
	id         string
	clients    map[*Connection]bool
	clientsMu  sync.RWMutex
	logger     *zerolog.Logger
	metadata   map[string]any
	metadataMu sync.RWMutex
}

// NewRoom creates a new room
func NewRoom(id string, logger *zerolog.Logger) *Room {
	return &Room{
		id:       id,
		clients:  make(map[*Connection]bool),
		logger:   logger,
		metadata: make(map[string]any),
	}
}

// ID returns the room ID
func (r *Room) ID() string {
	return r.id
}

// Join adds a connection to the room
func (r *Room) Join(conn *Connection) bool {
	r.clientsMu.Lock()

	if r.clients[conn] {
		r.clientsMu.Unlock()
		return false // Already in room
	}

	r.clients[conn] = true
	conn.Subscribe(r.id)
	r.clientsMu.Unlock()

	// Log subscription (after unlock to avoid lock ordering issues)
	r.logger.Info().
		Str("room", r.id).
		Strs("subscriptions", conn.GetSubscriptions()).
		Msg("Connection joined room")

	return true
}

// Leave removes a connection from the room
func (r *Room) Leave(conn *Connection) bool {
	r.clientsMu.Lock()
	defer r.clientsMu.Unlock()

	if !r.clients[conn] {
		return false // Not in room
	}

	delete(r.clients, conn)
	conn.Unsubscribe(r.id)
	return true
}

// Size returns the number of connections in the room
func (r *Room) Size() int {
	r.clientsMu.RLock()
	defer r.clientsMu.RUnlock()
	return len(r.clients)
}

// Broadcast sends a message to all connections in the room
func (r *Room) Broadcast(message any) {
	r.clientsMu.RLock()
	clients := make([]*Connection, 0, len(r.clients))
	for conn := range r.clients {
		clients = append(clients, conn)
	}
	r.clientsMu.RUnlock()

	// Send to all clients (outside of lock to avoid deadlock)
	for _, conn := range clients {
		if err := conn.WriteJSON(message); err != nil {
			r.logger.Debug().
				Str("room", r.id).
				Err(err).
				Msg("Failed to send message to client in room")
		}
	}
}

// BroadcastExcluding sends a message to all connections except the specified one
func (r *Room) BroadcastExcluding(message any, exclude *Connection) {
	r.clientsMu.RLock()
	clients := make([]*Connection, 0, len(r.clients))
	for conn := range r.clients {
		if conn != exclude {
			clients = append(clients, conn)
		}
	}
	r.clientsMu.RUnlock()

	// Send to all clients (outside of lock)
	for _, conn := range clients {
		if err := conn.WriteJSON(message); err != nil {
			r.logger.Debug().
				Str("room", r.id).
				Err(err).
				Msg("Failed to send message to client in room")
		}
	}
}

// GetClients returns a snapshot of all clients in the room
func (r *Room) GetClients() []*Connection {
	r.clientsMu.RLock()
	defer r.clientsMu.RUnlock()

	clients := make([]*Connection, 0, len(r.clients))
	for conn := range r.clients {
		clients = append(clients, conn)
	}
	return clients
}

// SetMetadata sets room metadata
func (r *Room) SetMetadata(key string, value any) {
	r.metadataMu.Lock()
	defer r.metadataMu.Unlock()
	r.metadata[key] = value
}

// GetMetadata gets room metadata
func (r *Room) GetMetadata(key string) (any, bool) {
	r.metadataMu.RLock()
	defer r.metadataMu.RUnlock()
	value, ok := r.metadata[key]
	return value, ok
}
