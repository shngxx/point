package usecase

import (
	"context"
	"fmt"

	"github.com/shngxx/point/internal/domain/point"
)

// GetPointUC implements the use case: getting point information
type GetPointUC struct {
	pointRepository point.PointRepository
}

// NewGetPointUC creates a new use case for getting point information
func NewGetPointUC(repository point.PointRepository) *GetPointUC {
	return &GetPointUC{
		pointRepository: repository,
	}
}

// PointInfo contains information about a point
type PointInfo struct {
	ID    int          `json:"id"`
	Point *point.Point `json:"point"`
}

// GetPoint executes the use case: gets point information by ID
func (u *GetPointUC) GetPoint(ctx context.Context, id int) (*PointInfo, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid point id: %d", id)
	}

	p, err := u.pointRepository.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get point: %w", err)
	}

	return &PointInfo{
		ID:    id,
		Point: &point.Point{X: p.X, Y: p.Y},
	}, nil
}
