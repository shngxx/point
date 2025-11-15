package usecase

import (
	"github.com/shngxx/point/internal/domain/point"
)

// GetPointUC реализует сценарий использования: получение информации о точке
type GetPointUC struct {
	pointRepository point.PointRepository
}

// NewGetPointUC создает новый usecase для получения информации о точке
func NewGetPointUC(repository point.PointRepository) *GetPointUC {
	return &GetPointUC{
		pointRepository: repository,
	}
}

// PointInfo содержит информацию о точке
type PointInfo struct {
	ID int         `json:"id"`
	Point *point.Point `json:"point"`
}

// Execute выполняет сценарий: получает информацию о точке по ID
func (u *GetPointUC) Execute(id int) *PointInfo {
	p := u.pointRepository.Get(id)
	return &PointInfo{
		ID:    id,
		Point: &point.Point{X: p.X, Y: p.Y},
	}
}

