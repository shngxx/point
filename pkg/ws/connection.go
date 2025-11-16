package ws

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/gofiber/websocket/v2"
	"github.com/rs/zerolog"
)

// Connection wraps websocket.Conn with enhanced functionality
type Connection struct {
	conn   *websocket.Conn
	logger *zerolog.Logger

	// Metadata storage
	metadata   map[string]any
	metadataMu sync.RWMutex

	// Subscription tracking (rooms this connection is in)
	rooms   map[string]bool
	roomsMu sync.RWMutex

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Message channels
	readChan  chan []byte
	writeChan chan any
	errorChan chan error

	// Connection state
	closed   bool
	closedMu sync.RWMutex
}

// NewConnection creates a new Connection wrapper
func NewConnection(conn *websocket.Conn, logger *zerolog.Logger) *Connection {
	ctx, cancel := context.WithCancel(context.Background())

	return &Connection{
		conn:      conn,
		logger:    logger,
		metadata:  make(map[string]any),
		rooms:     make(map[string]bool),
		ctx:       ctx,
		cancel:    cancel,
		readChan:  make(chan []byte, 256),
		writeChan: make(chan any, 256),
		errorChan: make(chan error, 1),
	}
}

// Start starts the connection handlers (read and write goroutines)
func (c *Connection) Start(ctx context.Context) {
	// Start read goroutine
	go c.readLoop()

	// Start write goroutine
	go c.writeLoop()
}

// readLoop continuously reads messages from the WebSocket connection
func (c *Connection) readLoop() {
	defer close(c.readChan)
	defer close(c.errorChan)

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error().Err(err).Msg("WebSocket read error")
				}
				c.errorChan <- err
				return
			}

			select {
			case c.readChan <- message:
			case <-c.ctx.Done():
				return
			}
		}
	}
}

// writeLoop continuously writes messages to the WebSocket connection
func (c *Connection) writeLoop() {
	for {
		select {
		case <-c.ctx.Done():
			return
		case msg := <-c.writeChan:
			if c.isClosed() {
				return
			}

			var data []byte
			var err error

			switch v := msg.(type) {
			case []byte:
				data = v
			case string:
				data = []byte(v)
			default:
				data, err = json.Marshal(msg)
				if err != nil {
					c.logger.Error().Err(err).Msg("Failed to marshal message")
					continue
				}
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, data); err != nil {
				c.logger.Error().Err(err).Msg("WebSocket write error")
				return
			}
		}
	}
}

// ReadJSON reads a JSON message from the connection
func (c *Connection) ReadJSON(v any) error {
	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case message, ok := <-c.readChan:
		if !ok {
			return websocket.ErrCloseSent
		}
		return json.Unmarshal(message, v)
	case err := <-c.errorChan:
		return err
	}
}

// WriteJSON writes a JSON message to the connection
func (c *Connection) WriteJSON(v any) error {
	if c.isClosed() {
		return websocket.ErrCloseSent
	}

	select {
	case <-c.ctx.Done():
		return c.ctx.Err()
	case c.writeChan <- v:
		return nil
	default:
		// Channel is full, message dropped
		c.logger.Warn().Msg("Write channel full, message dropped")
		return nil
	}
}

// Close closes the connection
func (c *Connection) Close() error {
	c.closedMu.Lock()
	if c.closed {
		c.closedMu.Unlock()
		return nil
	}
	c.closed = true
	c.closedMu.Unlock()

	c.cancel()
	return c.conn.Close()
}

// isClosed checks if the connection is closed
func (c *Connection) isClosed() bool {
	c.closedMu.RLock()
	defer c.closedMu.RUnlock()
	return c.closed
}

// Context returns the connection's context
func (c *Connection) Context() context.Context {
	return c.ctx
}

// SetMetadata sets a metadata value
func (c *Connection) SetMetadata(key string, value any) {
	c.metadataMu.Lock()
	defer c.metadataMu.Unlock()
	c.metadata[key] = value
}

// GetMetadata gets a metadata value
func (c *Connection) GetMetadata(key string) (any, bool) {
	c.metadataMu.RLock()
	defer c.metadataMu.RUnlock()
	value, ok := c.metadata[key]
	return value, ok
}

// Subscribe adds the connection to a room
func (c *Connection) Subscribe(roomID string) {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()
	c.rooms[roomID] = true
}

// Unsubscribe removes the connection from a room
func (c *Connection) Unsubscribe(roomID string) {
	c.roomsMu.Lock()
	defer c.roomsMu.Unlock()
	delete(c.rooms, roomID)
}

// GetSubscriptions returns all room IDs the connection is subscribed to
func (c *Connection) GetSubscriptions() []string {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()

	rooms := make([]string, 0, len(c.rooms))
	for roomID := range c.rooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}

// IsSubscribed checks if the connection is subscribed to a room
func (c *Connection) IsSubscribed(roomID string) bool {
	c.roomsMu.RLock()
	defer c.roomsMu.RUnlock()
	return c.rooms[roomID]
}

// Conn returns the underlying websocket.Conn (for advanced use cases)
func (c *Connection) Conn() *websocket.Conn {
	return c.conn
}
