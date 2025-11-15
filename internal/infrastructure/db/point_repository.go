package db

import (
	"github.com/shngxx/point/internal/domain/point"
)

// PointRepository реализует интерфейс domain.PointRepository
type PointRepository struct {
	// TODO: в будущем здесь будет настоящая БД
	point *point.Point
}

// NewPointRepository создает новый репозиторий
func NewPointRepository() *PointRepository {
	// Создаем точку с дефолтными значениями
	p := point.NewPoint(0, 0)
	return &PointRepository{
		point: p,
	}
}

// Get возвращает точку по идентификатору
func (r *PointRepository) Get(id int) *point.Point {
	// TODO: в будущем будет запрос к БД по id
	// Пока возвращаем одну и ту же точку
	return r.point
}

// Save сохраняет точку по идентификатору
func (r *PointRepository) Save(id int, p *point.Point) {
	// TODO: в будущем будет сохранение в БД
	// Пока просто обновляем точку в памяти
	r.point.X = p.X
	r.point.Y = p.Y
}
