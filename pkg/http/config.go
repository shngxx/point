package http

import (
	"fmt"
	"time"
)

// ServerConfig defines the interface for server configuration
// Implementations should provide server settings without binding to specific config libraries
type ServerConfig interface {
	// GetAddress returns the server address (e.g., ":8080" or "localhost:8080")
	GetAddress() string

	// GetPort returns the server port
	GetPort() int

	// GetReadTimeout returns the read timeout duration
	GetReadTimeout() time.Duration

	// GetWriteTimeout returns the write timeout duration
	GetWriteTimeout() time.Duration

	// GetIdleTimeout returns the idle timeout duration
	GetIdleTimeout() time.Duration

	// GetShutdownTimeout returns the graceful shutdown timeout duration
	GetShutdownTimeout() time.Duration
}

// Config represents server configuration that can be loaded via pkg/config
// Use this type with config.Load or config.LoadSection to load from YAML
type Config struct {
	Host            string `koanf:"host"`
	Port            int    `koanf:"port"`
	ReadTimeout     int    `koanf:"readTimeout"`     // in seconds
	WriteTimeout    int    `koanf:"writeTimeout"`    // in seconds
	IdleTimeout     int    `koanf:"idleTimeout"`     // in seconds (optional, default: 120)
	ShutdownTimeout int    `koanf:"shutdownTimeout"` // in seconds (optional, default: 30)
}

// GetAddress returns the server address
func (c Config) GetAddress() string {
	if c.Host != "" && c.Port > 0 {
		return fmt.Sprintf("%s:%d", c.Host, c.Port)
	}
	if c.Port > 0 {
		return fmt.Sprintf(":%d", c.Port)
	}
	return ":8080"
}

// GetPort returns the server port
func (c Config) GetPort() int {
	if c.Port > 0 {
		return c.Port
	}
	return 8080
}

// GetReadTimeout returns the read timeout
func (c Config) GetReadTimeout() time.Duration {
	if c.ReadTimeout > 0 {
		return time.Duration(c.ReadTimeout) * time.Second
	}
	return 10 * time.Second
}

// GetWriteTimeout returns the write timeout
func (c Config) GetWriteTimeout() time.Duration {
	if c.WriteTimeout > 0 {
		return time.Duration(c.WriteTimeout) * time.Second
	}
	return 10 * time.Second
}

// GetIdleTimeout returns the idle timeout
func (c Config) GetIdleTimeout() time.Duration {
	if c.IdleTimeout > 0 {
		return time.Duration(c.IdleTimeout) * time.Second
	}
	return 120 * time.Second
}

// GetShutdownTimeout returns the shutdown timeout
func (c Config) GetShutdownTimeout() time.Duration {
	if c.ShutdownTimeout > 0 {
		return time.Duration(c.ShutdownTimeout) * time.Second
	}
	return 30 * time.Second
}

// DefaultConfig provides default server configuration values
type DefaultConfig struct {
	Address         string
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// GetAddress returns the server address
func (c *DefaultConfig) GetAddress() string {
	if c.Address != "" {
		return c.Address
	}
	return ":8080"
}

// GetPort returns the server port
func (c *DefaultConfig) GetPort() int {
	if c.Port > 0 {
		return c.Port
	}
	return 8080
}

// GetReadTimeout returns the read timeout
func (c *DefaultConfig) GetReadTimeout() time.Duration {
	if c.ReadTimeout > 0 {
		return c.ReadTimeout
	}
	return 10 * time.Second
}

// GetWriteTimeout returns the write timeout
func (c *DefaultConfig) GetWriteTimeout() time.Duration {
	if c.WriteTimeout > 0 {
		return c.WriteTimeout
	}
	return 10 * time.Second
}

// GetIdleTimeout returns the idle timeout
func (c *DefaultConfig) GetIdleTimeout() time.Duration {
	if c.IdleTimeout > 0 {
		return c.IdleTimeout
	}
	return 120 * time.Second
}

// GetShutdownTimeout returns the shutdown timeout
func (c *DefaultConfig) GetShutdownTimeout() time.Duration {
	if c.ShutdownTimeout > 0 {
		return c.ShutdownTimeout
	}
	return 30 * time.Second
}
