package ws

import (
	"time"
)

// ManagerConfig defines the interface for WebSocket manager configuration
// Implementations should provide WebSocket settings without binding to specific config libraries
type ManagerConfig interface {
	// GetPingInterval returns the ping interval duration
	GetPingInterval() time.Duration

	// GetPongTimeout returns the pong timeout duration
	GetPongTimeout() time.Duration

	// GetReadBufferSize returns the read buffer size in bytes
	GetReadBufferSize() int

	// GetWriteBufferSize returns the write buffer size in bytes
	GetWriteBufferSize() int

	// GetMaxConnectionsPerRoom returns the maximum number of connections per room (0 = unlimited)
	GetMaxConnectionsPerRoom() int

	// GetShutdownTimeout returns the graceful shutdown timeout duration
	GetShutdownTimeout() time.Duration
}

// Config represents WebSocket manager configuration that can be loaded via pkg/config
// Use this type with config.Load or config.LoadSection to load from YAML
type Config struct {
	PingInterval          int `koanf:"pingInterval"`          // in seconds
	PongTimeout           int `koanf:"pongTimeout"`           // in seconds
	ReadBufferSize        int `koanf:"readBufferSize"`        // in bytes
	WriteBufferSize       int `koanf:"writeBufferSize"`       // in bytes
	MaxConnectionsPerRoom int `koanf:"maxConnectionsPerRoom"` // 0 = unlimited
	ShutdownTimeout       int `koanf:"shutdownTimeout"`       // in seconds
}

// GetPingInterval returns the ping interval
func (c *Config) GetPingInterval() time.Duration {
	if c.PingInterval > 0 {
		return time.Duration(c.PingInterval) * time.Second
	}
	return 60 * time.Second // Default: 60 seconds
}

// GetPongTimeout returns the pong timeout
func (c *Config) GetPongTimeout() time.Duration {
	if c.PongTimeout > 0 {
		return time.Duration(c.PongTimeout) * time.Second
	}
	return 10 * time.Second // Default: 10 seconds
}

// GetReadBufferSize returns the read buffer size
func (c *Config) GetReadBufferSize() int {
	if c.ReadBufferSize > 0 {
		return c.ReadBufferSize
	}
	return 4096 // Default: 4KB
}

// GetWriteBufferSize returns the write buffer size
func (c *Config) GetWriteBufferSize() int {
	if c.WriteBufferSize > 0 {
		return c.WriteBufferSize
	}
	return 4096 // Default: 4KB
}

// GetMaxConnectionsPerRoom returns the maximum connections per room
func (c *Config) GetMaxConnectionsPerRoom() int {
	return c.MaxConnectionsPerRoom // 0 = unlimited
}

// GetShutdownTimeout returns the shutdown timeout
func (c *Config) GetShutdownTimeout() time.Duration {
	if c.ShutdownTimeout > 0 {
		return time.Duration(c.ShutdownTimeout) * time.Second
	}
	return 30 * time.Second // Default: 30 seconds
}

// DefaultConfig provides default WebSocket manager configuration values
type DefaultConfig struct {
	PingInterval          time.Duration
	PongTimeout           time.Duration
	ReadBufferSize        int
	WriteBufferSize       int
	MaxConnectionsPerRoom int
	ShutdownTimeout       time.Duration
}

// GetPingInterval returns the ping interval
func (c *DefaultConfig) GetPingInterval() time.Duration {
	if c.PingInterval > 0 {
		return c.PingInterval
	}
	return 60 * time.Second
}

// GetPongTimeout returns the pong timeout
func (c *DefaultConfig) GetPongTimeout() time.Duration {
	if c.PongTimeout > 0 {
		return c.PongTimeout
	}
	return 10 * time.Second
}

// GetReadBufferSize returns the read buffer size
func (c *DefaultConfig) GetReadBufferSize() int {
	if c.ReadBufferSize > 0 {
		return c.ReadBufferSize
	}
	return 4096
}

// GetWriteBufferSize returns the write buffer size
func (c *DefaultConfig) GetWriteBufferSize() int {
	if c.WriteBufferSize > 0 {
		return c.WriteBufferSize
	}
	return 4096
}

// GetMaxConnectionsPerRoom returns the maximum connections per room
func (c *DefaultConfig) GetMaxConnectionsPerRoom() int {
	return c.MaxConnectionsPerRoom
}

// GetShutdownTimeout returns the shutdown timeout
func (c *DefaultConfig) GetShutdownTimeout() time.Duration {
	if c.ShutdownTimeout > 0 {
		return c.ShutdownTimeout
	}
	return 30 * time.Second
}
