package point

// Point represents a point on a plane with boundaries
type Point struct {
	X    int `json:"x"`
	Y    int `json:"y"`
	MaxX int `json:"-"`
	MaxY int `json:"-"`
}

const (
	// DefaultX is the default X coordinate
	DefaultX = 400
	// DefaultY is the default Y coordinate
	DefaultY = 300
	// DefaultMaxX is the default maximum X coordinate
	DefaultMaxX = 800
	// DefaultMaxY is the default maximum Y coordinate
	DefaultMaxY = 600
)

// NewPoint creates a new point with given coordinates and boundaries
// If x or y equals 0, default values are used
// If maxX or maxY equals 0, default boundaries are used
func NewPoint(x, y, maxX, maxY int) *Point {
	if x == 0 {
		x = DefaultX
	}
	if y == 0 {
		y = DefaultY
	}
	if maxX == 0 {
		maxX = DefaultMaxX
	}
	if maxY == 0 {
		maxY = DefaultMaxY
	}
	return &Point{
		X:    x,
		Y:    y,
		MaxX: maxX,
		MaxY: maxY,
	}
}

// Move moves the point by the specified offsets with boundary clamping
// Boundaries are checked using MaxX and MaxY from the point itself
func (p *Point) Move(dx, dy int) {
	p.X += dx
	p.Y += dy
	p.Clamp()
}

// Clamp limits coordinates to the boundaries defined in the point
func (p *Point) Clamp() {
	if p.X < 0 {
		p.X = 0
	}
	if p.X >= p.MaxX {
		p.X = p.MaxX - 1
	}
	if p.Y < 0 {
		p.Y = 0
	}
	if p.Y >= p.MaxY {
		p.Y = p.MaxY - 1
	}
}
