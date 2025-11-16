package ws

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/rs/zerolog"
	"github.com/shngxx/point/pkg/ws/hooks"
	"github.com/shngxx/point/pkg/ws/middleware"
)

// Manager represents the WebSocket connection manager
type Manager struct {
	config      ManagerConfig
	logger      *zerolog.Logger
	middleware  []middleware.Handler
	hookManager *hooks.Manager
	router      *Router

	// Connection management
	connections map[*Connection]bool
	connMu      sync.RWMutex

	// Room management
	rooms  map[string]*Room
	roomMu sync.RWMutex

	// Shutdown
	shutdown     chan struct{}
	shutdownOnce sync.Once
}

// NewManager creates a new WebSocket manager instance with the given options
func NewManager(opts ...Option) *Manager {
	nop := zerolog.Nop()
	m := &Manager{
		logger:      &nop,
		config:      &DefaultConfig{},
		connections: make(map[*Connection]bool),
		rooms:       make(map[string]*Room),
		shutdown:    make(chan struct{}),
		hookManager: hooks.NewManager(),
		router:      NewRouter(),
	}

	// Apply options
	for _, opt := range opts {
		opt(m)
	}

	return m
}

// NewManagerWithDefaults creates a new WebSocket manager with default middleware stack
// This is a convenience function that sets up Logger and Recovery middleware automatically
func NewManagerWithDefaults(l *zerolog.Logger) *Manager {
	return NewManager(
		WithLogger(l),
		WithMiddleware(
			middleware.Logger(l),
			middleware.Recovery(l),
		),
	)
}

// HandleConnection handles a new WebSocket connection
// This is the entry point for new connections from Fiber
func (m *Manager) HandleConnection(c *websocket.Conn) {
	// Check if manager is shutting down
	select {
	case <-m.shutdown:
		c.Close()
		return
	default:
	}

	// Create connection wrapper
	conn := NewConnection(c, m.logger)

	// Apply middleware
	for _, mw := range m.middleware {
		if err := mw(conn); err != nil {
			m.logger.Error().Err(err).Msg("Middleware error")
			conn.Close()
			return
		}
	}

	// Register connection
	m.connMu.Lock()
	m.connections[conn] = true
	m.connMu.Unlock()

	// Execute OnConnect hook
	if err := m.hookManager.Execute(hooks.OnConnect, conn); err != nil {
		m.logger.Error().Err(err).Msg("OnConnect hook failed")
		conn.Close()
		return
	}

	m.logger.Info().Msg("New WebSocket connection established")

	// Defer cleanup
	defer func() {
		// Execute OnDisconnect hook
		m.hookManager.Execute(hooks.OnDisconnect, conn)

		// Remove from all rooms
		m.leaveAllRooms(conn)

		// Unregister connection
		m.connMu.Lock()
		delete(m.connections, conn)
		m.connMu.Unlock()

		conn.Close()
		m.logger.Info().Msg("WebSocket connection closed")
	}()

	// Start connection handlers
	conn.Start(context.Background())

	// Message handling loop
	m.handleMessages(conn)
}

// handleMessages handles incoming messages from a connection
func (m *Manager) handleMessages(conn *Connection) {
	for {
		select {
		case <-m.shutdown:
			return
		case <-conn.Context().Done():
			return
		default:
			var msg Message
			if err := conn.ReadJSON(&msg); err != nil {
				// Check if it's a connection close error
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					return
				}
				// For JSON parse errors, log and continue (might be ping/pong or empty message)
				if _, ok := err.(*json.SyntaxError); ok || err.Error() == "unexpected end of JSON input" {
					m.logger.Debug().Err(err).Msg("Invalid JSON message received, ignoring")
					continue
				}
				// For other errors, close connection
				return
			}

			// Skip empty messages
			if msg.Action == "" && msg.Type == "" {
				continue
			}

			// Execute OnMessage hook
			if err := m.hookManager.Execute(hooks.OnMessage, conn, &msg); err != nil {
				m.logger.Error().Err(err).Msg("OnMessage hook failed")
				continue
			}

			// Route message
			if err := m.router.Route(conn, &msg); err != nil {
				m.logger.Error().Err(err).Msg("Message routing error")
				// Send error response to client
				errorMsg := map[string]any{
					"error": err.Error(),
				}
				conn.WriteJSON(errorMsg)
			}
		}
	}
}

// leaveAllRooms removes connection from all rooms
func (m *Manager) leaveAllRooms(conn *Connection) {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	for roomID, room := range m.rooms {
		room.Leave(conn)
		// Cleanup empty rooms
		if room.Size() == 0 {
			delete(m.rooms, roomID)
		}
	}
}

// GetOrCreateRoom gets an existing room or creates a new one
func (m *Manager) GetOrCreateRoom(roomID string) *Room {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room, exists := m.rooms[roomID]
	if !exists {
		room = NewRoom(roomID, m.logger)
		m.rooms[roomID] = room
	}

	return room
}

// GetRoom gets an existing room
func (m *Manager) GetRoom(roomID string) (*Room, bool) {
	m.roomMu.RLock()
	defer m.roomMu.RUnlock()
	room, ok := m.rooms[roomID]
	return room, ok
}

// JoinRoom adds a connection to a room
func (m *Manager) JoinRoom(conn *Connection, roomID string) error {
	// Check max connections per room
	if maxConn := m.config.GetMaxConnectionsPerRoom(); maxConn > 0 {
		room, exists := m.GetRoom(roomID)
		if exists && room.Size() >= maxConn {
			return &Error{Code: "ROOM_FULL", Message: "Room is full"}
		}
	}

	room := m.GetOrCreateRoom(roomID)
	if room.Join(conn) {
		// Execute OnJoinRoom hook
		m.hookManager.Execute(hooks.OnJoinRoom, conn, roomID)
		m.logger.Debug().Str("room", roomID).Msg("Connection joined room")
	}
	return nil
}

// LeaveRoom removes a connection from a room
func (m *Manager) LeaveRoom(conn *Connection, roomID string) error {
	m.roomMu.Lock()
	defer m.roomMu.Unlock()

	room, exists := m.rooms[roomID]
	if !exists {
		return &Error{Code: "ROOM_NOT_FOUND", Message: "Room not found"}
	}

	if room.Leave(conn) {
		// Execute OnLeaveRoom hook
		m.hookManager.Execute(hooks.OnLeaveRoom, conn, roomID)
		m.logger.Debug().Str("room", roomID).Msg("Connection left room")

		// Cleanup empty rooms
		if room.Size() == 0 {
			delete(m.rooms, roomID)
		}
	}

	return nil
}

// BroadcastToRoom broadcasts a message to all connections in a room
func (m *Manager) BroadcastToRoom(roomID string, message any) error {
	m.roomMu.RLock()
	room, exists := m.rooms[roomID]
	m.roomMu.RUnlock()

	if !exists {
		return &Error{Code: "ROOM_NOT_FOUND", Message: "Room not found"}
	}

	room.Broadcast(message)
	return nil
}

// BroadcastToAll broadcasts a message to all connections
func (m *Manager) BroadcastToAll(message any) {
	m.connMu.RLock()
	connections := make([]*Connection, 0, len(m.connections))
	for conn := range m.connections {
		connections = append(connections, conn)
	}
	m.connMu.RUnlock()

	// Send to all connections
	for _, conn := range connections {
		if err := conn.WriteJSON(message); err != nil {
			m.logger.Debug().Err(err).Msg("Failed to broadcast to connection")
		}
	}
}

// SendToConnection sends a message to a specific connection
func (m *Manager) SendToConnection(conn *Connection, message any) error {
	return conn.WriteJSON(message)
}

// HandleMessage registers a message handler for a specific action
func (m *Manager) HandleMessage(action string, handler MessageHandler) {
	m.router.Handle(action, handler)
}

// GetConnectionCount returns the total number of connections
func (m *Manager) GetConnectionCount() int {
	m.connMu.RLock()
	defer m.connMu.RUnlock()
	return len(m.connections)
}

// GetRoomCount returns the total number of rooms
func (m *Manager) GetRoomCount() int {
	m.roomMu.RLock()
	defer m.roomMu.RUnlock()
	return len(m.rooms)
}

// Shutdown gracefully shuts down the manager
func (m *Manager) Shutdown() error {
	m.shutdownOnce.Do(func() {
		close(m.shutdown)

		// Close all connections with timeout
		ctx, cancel := context.WithTimeout(context.Background(), m.config.GetShutdownTimeout())
		defer cancel()

		done := make(chan struct{})
		go func() {
			m.connMu.RLock()
			for conn := range m.connections {
				conn.Close()
			}
			m.connMu.RUnlock()
			close(done)
		}()

		select {
		case <-done:
			// All connections closed
		case <-ctx.Done():
			m.logger.Warn().Msg("Shutdown timeout reached, some connections may not have closed gracefully")
		}

		m.logger.Info().Msg("WebSocket manager shutdown completed")
	})

	return nil
}
