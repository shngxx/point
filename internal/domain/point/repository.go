package point

import "context"

// PointRepository определяет интерфейс репозитория для работы с точкой
type PointRepository interface {
	// Get возвращает точку по идентификатору
	Get(ctx context.Context, id int) (*Point, error)

	// Save сохраняет точку по идентификатору
	Save(ctx context.Context, id int, p *Point) error
}
