package usecase

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/shngxx/point/internal/domain/point"
)

// MoveCommand represents a command to move a point
type MoveCommand struct {
	ID int
	DX int
	DY int
}

// MovePointConfig contains configuration for MovePointUC
type MovePointConfig struct {
	BatchInterval time.Duration // Batch processing interval (~60 FPS)
	SaveInterval  time.Duration // Position save interval
}

// MovePointUC implements the use case: step-by-step point movement
type MovePointUC struct {
	pointRepository point.PointRepository
	logger          *zerolog.Logger
	config          MovePointConfig
}

// NewMovePointUC creates a new use case for step-by-step point movement
func NewMovePointUC(
	repository point.PointRepository,
	logger *zerolog.Logger,
	config MovePointConfig,
) *MovePointUC {
	return &MovePointUC{
		pointRepository: repository,
		logger:          logger,
		config:          config,
	}
}

// ClientSession represents a client session with a separate command channel
type ClientSession struct {
	moveChan     chan MoveCommand
	positionChan chan *point.Point
}

// PositionChan returns a channel for receiving position updates
func (s *ClientSession) PositionChan() <-chan *point.Point {
	return s.positionChan
}

// Init starts a goroutine to process point movement
// Called once when WebSocket connection is activated
// Returns a client session with channels for commands and position updates
func (u *MovePointUC) Init(ctx context.Context, id int) *ClientSession {
	// Create a separate command channel for this client
	moveChan := make(chan MoveCommand, 50)
	positionChan := make(chan *point.Point, 5)

	session := &ClientSession{
		moveChan:     moveChan,
		positionChan: positionChan,
	}

	go u.processMoves(ctx, id, session)
	return session
}

// Push adds a move command to the client channel
func (s *ClientSession) Push(cmd MoveCommand) {
	select {
	case s.moveChan <- cmd:
	default:
		// Channel is full, ignore command
	}
}

// processMoves processes move commands in an infinite loop
// session - client session with channels for commands and position updates
func (u *MovePointUC) processMoves(ctx context.Context, id int, session *ClientSession) {
	ticker := time.NewTicker(u.config.SaveInterval)
	defer ticker.Stop()
	defer close(session.positionChan)
	defer close(session.moveChan)

	// Timer for batching commands
	batchTicker := time.NewTicker(u.config.BatchInterval)
	defer batchTicker.Stop()

	var pendingCommands []MoveCommand
	lastSentPos := &point.Point{X: -1, Y: -1} // For tracking changes

	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-session.moveChan:
			// Accumulate commands for batching
			pendingCommands = append(pendingCommands, cmd)
		case <-batchTicker.C:
			// Process accumulated commands in batch
			if len(pendingCommands) > 0 {
				if err := u.processBatch(ctx, id, session, pendingCommands, lastSentPos); err != nil {
					u.logger.Error().Err(err).Msg("Error processing batch")
					pendingCommands = pendingCommands[:0]
					continue
				}
				pendingCommands = pendingCommands[:0] // Clear slice
			}
		case <-ticker.C:
			// Periodically save point position
			if err := u.savePoint(ctx, id); err != nil {
				u.logger.Error().Err(err).Msg("Error saving point")
				continue
			}
		}
	}
}

// processBatch processes a batch of move commands
func (u *MovePointUC) processBatch(ctx context.Context, id int, session *ClientSession, commands []MoveCommand, lastSentPos *point.Point) error {
	p, err := u.pointRepository.Get(ctx, id)
	if err != nil {
		return err
	}

	oldX, oldY := p.X, p.Y

	// Apply all commands sequentially
	// Boundaries are checked inside Move method from domain level
	for _, cmd := range commands {
		p.Move(cmd.DX, cmd.DY)
	}
	commandCount := len(commands)

	// Save updated position
	if err := u.pointRepository.Save(ctx, id, p); err != nil {
		return err
	}

	// Send update only if position changed
	if p.X != lastSentPos.X || p.Y != lastSentPos.Y {
		lastSentPos.X = p.X
		lastSentPos.Y = p.Y

		// Log point movement
		u.logger.Debug().
			Int("id", id).
			Int("oldX", oldX).
			Int("newX", p.X).
			Int("oldY", oldY).
			Int("newY", p.Y).
			Int("commands", commandCount).
			Msg("Point moved")

		select {
		case session.positionChan <- &point.Point{X: p.X, Y: p.Y}:
		default:
			// Channel is full, ignore
		}
	}

	return nil
}

// savePoint saves the current point position
func (u *MovePointUC) savePoint(ctx context.Context, id int) error {
	p, err := u.pointRepository.Get(ctx, id)
	if err != nil {
		return err
	}

	if err := u.pointRepository.Save(ctx, id, p); err != nil {
		return err
	}

	u.logger.Debug().
		Int("id", id).
		Int("x", p.X).
		Int("y", p.Y).
		Msg("Point saved successfully")

	return nil
}
