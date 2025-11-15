package point

// Point представляет точку на плоскости
type Point struct {
	ID int `json:"id"`
	X  int `json:"x"`
	Y  int `json:"y"`
}

const (
	// DefaultX дефолтная координата X
	DefaultX = 400
	// DefaultY дефолтная координата Y
	DefaultY = 300
	// DefaultMaxX дефолтная максимальная координата X
	DefaultMaxX = 800
	// DefaultMaxY дефолтная максимальная координата Y
	DefaultMaxY = 600
)

// NewPoint создает новую точку с заданными координатами
// Если x или y равны 0, используются дефолтные значения
func NewPoint(x, y int) *Point {
	if x == 0 {
		x = DefaultX
	}
	if y == 0 {
		y = DefaultY
	}
	return &Point{
		X: x,
		Y: y,
	}
}

// Move перемещает точку на указанные смещения с ограничением границами
// Использует дефолтные границы из домена
func (p *Point) Move(dx, dy int) {
	p.X += dx
	p.Y += dy
	p.Clamp(DefaultMaxX, DefaultMaxY)
}

// Teleport устанавливает новые координаты точки с ограничением границами
// Использует дефолтные границы из домена
func (p *Point) Teleport(x, y int) {
	p.X = x
	p.Y = y
	p.Clamp(DefaultMaxX, DefaultMaxY)
}

// Validate проверяет, что координаты находятся в допустимых пределах
func (p *Point) Validate(maxX, maxY int) bool {
	return p.X >= 0 && p.X < maxX && p.Y >= 0 && p.Y < maxY
}

// Clamp ограничивает координаты заданными пределами
func (p *Point) Clamp(maxX, maxY int) {
	if p.X < 0 {
		p.X = 0
	}
	if p.X >= maxX {
		p.X = maxX - 1
	}
	if p.Y < 0 {
		p.Y = 0
	}
	if p.Y >= maxY {
		p.Y = maxY - 1
	}
}
