package point

// PointRepository определяет интерфейс репозитория для работы с точкой
type PointRepository interface {
	// Get возвращает точку по идентификатору
	Get(id int) *Point

	// Save сохраняет точку по идентификатору
	Save(id int, p *Point)
}
