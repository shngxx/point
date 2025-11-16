package ws

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/rs/zerolog"
	"github.com/shngxx/point/internal/domain/point"
	"github.com/shngxx/point/internal/usecase"
	wsmanager "github.com/shngxx/point/pkg/ws"
)

// GetPointService defines the interface for getting point information
type GetPointService interface {
	GetPoint(ctx context.Context, id int) (*usecase.PointInfo, error)
}

// MovePointService defines the interface for point movement
type MovePointService interface {
	// Init starts a goroutine to process point movement
	// Returns a client session with channels for commands and position updates
	Init(ctx context.Context, id int) *usecase.ClientSession
}

// MoveMessage represents a message from the client to move the point
type MoveMessage struct {
	DX int `json:"dx,omitempty"`
	DY int `json:"dy,omitempty"`
}

// PositionMessage represents a position message for the client
type PositionMessage struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Handler handles WebSocket connections using pkg/ws.Manager
type Handler struct {
	manager          *wsmanager.Manager
	getPointService  GetPointService
	movePointService MovePointService
	logger           *zerolog.Logger
	sessions         map[*wsmanager.Connection]*usecase.ClientSession
	sessionsMu       sync.RWMutex
}

// NewHandler creates a new WebSocket handler
func NewHandler(
	manager *wsmanager.Manager,
	getPointService GetPointService,
	movePointService MovePointService,
	logger *zerolog.Logger,
) *Handler {
	h := &Handler{
		manager:          manager,
		getPointService:  getPointService,
		movePointService: movePointService,
		logger:           logger,
		sessions:         make(map[*wsmanager.Connection]*usecase.ClientSession),
	}

	// Register message handlers
	h.registerHandlers()

	return h
}

// registerHandlers registers message handlers with the manager
func (h *Handler) registerHandlers() {
	// Handle move commands
	h.manager.HandleMessage("move", h.handleMove)
}

// handleMove handles move commands from the client
func (h *Handler) handleMove(conn *wsmanager.Connection, msg *wsmanager.Message) error {
	var moveMsg MoveMessage

	// Try to parse from data field first
	if len(msg.Data) > 0 {
		if err := json.Unmarshal(msg.Data, &moveMsg); err != nil {
			// If data field parsing fails, try parsing the whole message
			// This handles the case when frontend sends {action: 'move', dx: ..., dy: ...}
			var fullMsg struct {
				Action string `json:"action"`
				DX     int    `json:"dx"`
				DY     int    `json:"dy"`
			}
			// Re-parse the original message
			msgBytes, _ := json.Marshal(msg)
			if err := json.Unmarshal(msgBytes, &fullMsg); err == nil {
				moveMsg.DX = fullMsg.DX
				moveMsg.DY = fullMsg.DY
			} else {
				return err
			}
		}
	} else {
		// If data is empty, try to get dx/dy from message directly
		// This is a fallback for messages like {action: 'move', dx: 10, dy: 0}
		var fullMsg map[string]any
		msgBytes, _ := json.Marshal(msg)
		if err := json.Unmarshal(msgBytes, &fullMsg); err == nil {
			if dx, ok := fullMsg["dx"].(float64); ok {
				moveMsg.DX = int(dx)
			}
			if dy, ok := fullMsg["dy"].(float64); ok {
				moveMsg.DY = int(dy)
			}
		}
	}

	// Get or create session for this connection
	session := h.getOrCreateSession(conn)

	// Get point ID from connection metadata or use default
	pointID := 1
	if pointIDVal, ok := conn.GetMetadata("point_id"); ok {
		if id, ok := pointIDVal.(int); ok {
			pointID = id
		}
	}

	// If there's a move command, add it to the client channel
	if moveMsg.DX != 0 || moveMsg.DY != 0 {
		session.Push(usecase.MoveCommand{
			ID: pointID,
			DX: moveMsg.DX,
			DY: moveMsg.DY,
		})
	}

	return nil
}

// getOrCreateSession gets or creates a session for a connection
func (h *Handler) getOrCreateSession(conn *wsmanager.Connection) *usecase.ClientSession {
	h.sessionsMu.Lock()
	defer h.sessionsMu.Unlock()

	session, exists := h.sessions[conn]
	if !exists {
		// Get point ID from connection metadata or use default
		pointID := 1
		if pointIDVal, ok := conn.GetMetadata("point_id"); ok {
			if id, ok := pointIDVal.(int); ok {
				pointID = id
			}
		}

		// Initialize point movement processing
		session = h.movePointService.Init(conn.Context(), pointID)
		h.sessions[conn] = session

		// Start goroutine to send position updates
		go h.sendPositionUpdates(conn, session, pointID)
	}

	return session
}

// sendPositionUpdates sends position updates from the session to the connection
func (h *Handler) sendPositionUpdates(conn *wsmanager.Connection, session *usecase.ClientSession, pointID int) {
	roomID := "point_" + strconv.Itoa(pointID)

	// Join room for this point
	if err := h.manager.JoinRoom(conn, roomID); err != nil {
		h.logger.Error().Str("room", roomID).Err(err).Msg("Failed to join room")
	}

	for {
		select {
		case <-conn.Context().Done():
			// Cleanup session
			h.sessionsMu.Lock()
			delete(h.sessions, conn)
			h.sessionsMu.Unlock()
			return
		case pos := <-session.PositionChan():
			if pos == nil {
				// Channel closed
				return
			}
			h.sendPosition(conn, pos)
		}
	}
}

// sendPosition sends position to a connection
func (h *Handler) sendPosition(conn *wsmanager.Connection, pos *point.Point) {
	msg := PositionMessage{
		X: pos.X,
		Y: pos.Y,
	}
	if err := conn.WriteJSON(msg); err != nil {
		h.logger.Error().Err(err).Msg("WebSocket send error")
	}
}

// BroadcastPosition sends position to all connected clients for a specific point
// Used for managing point from backend
func (h *Handler) BroadcastPosition(ctx context.Context, pointID int) {
	pointInfo, err := h.getPointService.GetPoint(ctx, pointID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Error getting point for broadcast")
		return
	}

	roomID := "point_" + strconv.Itoa(pointID)
	msg := PositionMessage{
		X: pointInfo.Point.X,
		Y: pointInfo.Point.Y,
	}

	if err := h.manager.BroadcastToRoom(roomID, msg); err != nil {
		h.logger.Error().Str("room", roomID).Err(err).Msg("Error broadcasting position")
	}
}

// Manager returns the underlying WebSocket manager
func (h *Handler) Manager() *wsmanager.Manager {
	return h.manager
}
