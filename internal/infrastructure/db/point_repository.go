package db

import (
	"context"
	"fmt"
	"sync"

	"github.com/shngxx/point/internal/domain/point"
)

// PointRepository implements the domain.PointRepository interface
type PointRepository struct {
	mu     sync.RWMutex
	points map[int]*point.Point
}

// NewPointRepository creates a new repository
func NewPointRepository() *PointRepository {
	// Initialize with default point
	points := make(map[int]*point.Point)
	// Create default point with ID 1 and boundaries
	points[1] = point.NewPoint(0, 0, 0, 0)
	return &PointRepository{
		points: points,
	}
}

// Get returns a point by identifier
func (r *PointRepository) Get(ctx context.Context, id int) (*point.Point, error) {
	// Check context
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	// TODO: in the future this will be a database query by id
	// For now, return the point from memory or create a default one
	p, exists := r.points[id]
	if !exists {
		// Return default point if not found (use boundaries from first point if exists)
		if len(r.points) > 0 {
			// Use boundaries from existing point
			for _, existingPoint := range r.points {
				p = point.NewPoint(0, 0, existingPoint.MaxX, existingPoint.MaxY)
				break
			}
		} else {
			// Use default boundaries
			p = point.NewPoint(0, 0, 0, 0)
		}
	}

	// Create a copy for safety
	return &point.Point{
		X:    p.X,
		Y:    p.Y,
		MaxX: p.MaxX,
		MaxY: p.MaxY,
	}, nil
}

// Save saves a point by identifier
func (r *PointRepository) Save(ctx context.Context, id int, p *point.Point) error {
	// Check context
	if ctx.Err() != nil {
		return ctx.Err()
	}

	if p == nil {
		return fmt.Errorf("point cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// TODO: in the future this will be saved to database
	// For now, update the point in memory
	if r.points[id] == nil {
		// Create new point with boundaries from existing point or defaults
		if len(r.points) > 0 {
			for _, existingPoint := range r.points {
				r.points[id] = point.NewPoint(p.X, p.Y, existingPoint.MaxX, existingPoint.MaxY)
				return nil
			}
		}
		r.points[id] = point.NewPoint(p.X, p.Y, 0, 0)
		return nil
	}
	r.points[id].X = p.X
	r.points[id].Y = p.Y
	// Preserve boundaries
	r.points[id].MaxX = p.MaxX
	r.points[id].MaxY = p.MaxY

	return nil
}
