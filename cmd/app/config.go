package main

import (
	"time"

	"github.com/shngxx/point/pkg/http"
	applog "github.com/shngxx/point/pkg/log"
)

// AppConfig contains all application configuration
type AppConfig struct {
	Server http.Config   `koanf:"server"`
	Logger applog.Config `koanf:"logger"`
	Point  PointConfig   `koanf:"point"`
}

// PointConfig contains point-related configuration
type PointConfig struct {
	MaxX          int `koanf:"maxX"`          // Maximum X coordinate (default: 800)
	MaxY          int `koanf:"maxY"`          // Maximum Y coordinate (default: 600)
	BatchInterval int `koanf:"batchInterval"` // Batch processing interval in milliseconds (~60 FPS, default: 16ms)
	SaveInterval  int `koanf:"saveInterval"`  // Save interval in seconds (default: 5s)
}

// BatchInterval returns batch interval as time.Duration
func (c *PointConfig) BatchIntervalDuration() time.Duration {
	if c.BatchInterval > 0 {
		return time.Duration(c.BatchInterval) * time.Millisecond
	}
	return 16 * time.Millisecond // Default ~60 FPS
}

// SaveIntervalDuration returns save interval as time.Duration
func (c *PointConfig) SaveIntervalDuration() time.Duration {
	if c.SaveInterval > 0 {
		return time.Duration(c.SaveInterval) * time.Second
	}
	return 5 * time.Second // Default
}

// MaxXValue returns max X coordinate with default fallback
func (c *PointConfig) MaxXValue() int {
	if c.MaxX > 0 {
		return c.MaxX
	}
	return 800 // Default
}

// MaxYValue returns max Y coordinate with default fallback
func (c *PointConfig) MaxYValue() int {
	if c.MaxY > 0 {
		return c.MaxY
	}
	return 600 // Default
}
