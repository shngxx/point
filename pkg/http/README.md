# HTTP Wrapper for Fiber

A production-ready, extensible HTTP wrapper over the [Fiber](https://github.com/gofiber/fiber) framework for Go microservices.

## Features

- 🚀 **Simple API**: Clean and intuitive interface
- 🔌 **Extensible**: Plugin-based architecture with middleware support
- 🛡️ **Production Ready**: Graceful shutdown, health checks, error handling
- 🔒 **Framework Agnostic**: All dependencies through interfaces
- 📦 **Zero External Dependencies**: Only depends on Fiber
- 🎯 **Type Safe**: Full type safety with Go generics where applicable

## Quick Start

### Minimal Example

```go
package main

import (
    "github.com/shngxx/point/pkg/http"
    "github.com/shngxx/point/pkg/http/middleware"
)

func main() {
    // Create server with default settings
    server := http.New(
        http.WithAddress(":8080"),
    )
    
    // Register routes
    server.GET("/", func(c *http.Context) error {
        return http.OK(c, map[string]string{
            "message": "Hello World",
        })
    })
    
    // Start server (blocking, with graceful shutdown)
    if err := server.Run(); err != nil {
        panic(err)
    }
}
```

### Advanced Example

```go
package main

import (
    "time"
    
    "github.com/shngxx/point/pkg/http"
    "github.com/shngxx/point/pkg/http/hooks"
    "github.com/shngxx/point/pkg/http/middleware"
    "github.com/shngxx/point/pkg/http/routing"
    httplogger "github.com/shngxx/point/pkg/http/logger"
)

func main() {
    // Custom logger (implement httplogger.Logger interface)
    logger := NewMyLogger()
    
    // Custom configuration
    cfg := &http.DefaultConfig{
        Address:         ":8080",
        ReadTimeout:     10 * time.Second,
        WriteTimeout:    10 * time.Second,
        ShutdownTimeout: 30 * time.Second,
    }
    
    // Create server
    server := http.New(
        http.WithConfig(cfg),
        http.WithLogger(logger),
        
        // Global middleware
        http.WithMiddleware(
            middleware.Recovery(),
            middleware.Logger(logger),
            middleware.RequestID(),
            middleware.CORS(middleware.CORSConfig{
                AllowOrigins: []string{"*"},
                AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
            }),
            middleware.Timeout(30 * time.Second),
            middleware.Security(),
        ),
    )
    
    // Lifecycle hooks
    server.AddHook(hooks.BeforeStart, func() error {
        logger.Info("Initializing connections...")
        return nil
    })
    
    server.AddHook(hooks.AfterStart, func() error {
        logger.Info("Server started successfully")
        return nil
    })
    
    server.AddHook(hooks.BeforeShutdown, func() error {
        logger.Info("Closing connections...")
        return nil
    })
    
    // Routes
    server.GET("/", handleHome)
    server.GET("/users/:id", handleGetUser)
    
    // Route groups
    server.Group("/api/v1", func(g *routing.Group) {
        g.GET("/users", handleListUsers)
        g.POST("/users", handleCreateUser)
        
        // Nested group with middleware
        admin := g.Group("/admin", middleware.Auth(authService))
        {
            admin.GET("/users", handleAdminListUsers)
            admin.DELETE("/users/:id", handleAdminDeleteUser)
        }
    })
    
    // Start server
    if err := server.Run(); err != nil {
        logger.Error("Server failed", httplogger.Field{Key: "error", Value: err.Error()})
    }
}
```

## Configuration

### Server Configuration

The server can be configured using the `ServerConfig` interface or the `DefaultConfig` struct:

```go
type ServerConfig interface {
    GetAddress() string
    GetPort() int
    GetReadTimeout() time.Duration
    GetWriteTimeout() time.Duration
    GetIdleTimeout() time.Duration
    GetShutdownTimeout() time.Duration
}
```

Example:

```go
cfg := &http.DefaultConfig{
    Address:         ":8080",
    Port:            8080,
    ReadTimeout:     10 * time.Second,
    WriteTimeout:    10 * time.Second,
    IdleTimeout:     120 * time.Second,
    ShutdownTimeout: 30 * time.Second,
}

server := http.New(http.WithConfig(cfg))
```

### Functional Options

The server uses the functional options pattern for configuration:

- `WithAddress(addr string)` - Set server address
- `WithLogger(logger Logger)` - Set custom logger
- `WithConfig(cfg ServerConfig)` - Set server configuration
- `WithMiddleware(mw ...middleware.Handler)` - Set global middleware
- `WithErrorHandler(handler ErrorHandler)` - Set custom error handler
- `WithHealthCheck(check func() error)` - Set health check function
- `WithValidator(validator Validator)` - Set custom validator

## Middleware

### Built-in Middleware

#### Recovery

Recovers from panics and returns a 500 error:

```go
server.Use(middleware.Recovery())
```

#### Logger

Logs HTTP requests:

```go
server.Use(middleware.Logger(logger))
```

#### Request ID

Generates and propagates request IDs:

```go
server.Use(middleware.RequestID())

// In handler:
requestID := middleware.GetRequestID(c)
```

#### CORS

Handles CORS requests:

```go
server.Use(middleware.CORS(middleware.CORSConfig{
    AllowOrigins:     []string{"https://example.com"},
    AllowMethods:     []string{"GET", "POST"},
    AllowHeaders:     []string{"Content-Type", "Authorization"},
    AllowCredentials: true,
    MaxAge:           3600,
}))
```

#### Security Headers

Sets security headers:

```go
server.Use(middleware.Security())

// Or with CSP:
server.Use(middleware.SecurityWithCSP("default-src 'self'"))
```

#### Timeout

Sets request timeout:

```go
server.Use(middleware.Timeout(30 * time.Second))
```

### Custom Middleware

Create custom middleware by implementing the `middleware.Handler` type:

```go
func MyMiddleware() middleware.Handler {
    return func(c *fiber.Ctx) error {
        // Before request
        // ...
        
        err := c.Next()
        
        // After request
        // ...
        
        return err
    }
}
```

## Routing

### Basic Routes

```go
server.GET("/users", handleListUsers)
server.POST("/users", handleCreateUser)
server.PUT("/users/:id", handleUpdateUser)
server.DELETE("/users/:id", handleDeleteUser)
server.PATCH("/users/:id", handlePatchUser)
```

### Route Groups

```go
// Simple group
server.Group("/api/v1", func(g *routing.Group) {
    g.GET("/users", handleListUsers)
    g.POST("/users", handleCreateUser)
})

// Group with middleware
server.Group("/admin", func(g *routing.Group) {
    g.Use(middleware.Auth(authService))
    g.GET("/dashboard", handleDashboard)
})

// Nested groups
server.Group("/api", func(g *routing.Group) {
    g.Group("/v1", func(v1 *routing.Group) {
        v1.GET("/users", handleListUsers)
    })
})
```

## Error Handling

### Default Error Handler

The server includes a default error handler that maps errors to HTTP status codes:

```go
// Custom error handler
type MyErrorHandler struct{}

func (h *MyErrorHandler) Handle(c *fiber.Ctx, err error) error {
    // Custom error handling logic
    return c.Status(500).JSON(map[string]string{
        "error": err.Error(),
    })
}

server := http.New(
    http.WithErrorHandler(&MyErrorHandler{}),
)
```

### Response Helpers

Use response helpers for standardized responses:

```go
// Success responses
http.OK(c, data)           // 200 OK
http.Created(c, data)      // 201 Created

// Error responses
http.BadRequest(c, err)    // 400 Bad Request
http.NotFound(c, "Not found") // 404 Not Found
http.InternalError(c, err)  // 500 Internal Server Error
```

## Health Checks

The server automatically registers health check endpoints:

- `GET /health` - Liveness probe (always returns 200)
- `GET /ready` - Readiness probe (can be customized)

```go
server := http.New(
    http.WithHealthCheck(func() error {
        // Check database, Redis, etc.
        if db.Ping() != nil {
            return errors.New("database unavailable")
        }
        return nil
    }),
)
```

## Lifecycle Hooks

Register hooks for server lifecycle events:

```go
server.AddHook(hooks.BeforeStart, func() error {
    // Initialize connections
    return nil
})

server.AddHook(hooks.AfterStart, func() error {
    // Server started
    return nil
})

server.AddHook(hooks.BeforeShutdown, func() error {
    // Close connections
    return nil
})

server.AddHook(hooks.AfterShutdown, func() error {
    // Cleanup
    return nil
})
```

Available hook types:
- `hooks.BeforeStart` - Before server starts
- `hooks.AfterStart` - After server starts
- `hooks.BeforeShutdown` - Before shutdown begins
- `hooks.AfterShutdown` - After shutdown completes

## Logging

### Logger Interface

Implement the `logger.Logger` interface to use your own logger:

```go
type Logger interface {
    Info(msg string, fields ...Field)
    Error(msg string, fields ...Field)
    Warn(msg string, fields ...Field)
    Debug(msg string, fields ...Field)
}
```

### Example: Zap Logger Adapter

```go
type ZapLogger struct {
    logger *zap.Logger
}

func (l *ZapLogger) Info(msg string, fields ...httplogger.Field) {
    zapFields := make([]zap.Field, len(fields))
    for i, f := range fields {
        zapFields[i] = zap.Any(f.Key, f.Value)
    }
    l.logger.Info(msg, zapFields...)
}

// ... implement other methods

server := http.New(
    http.WithLogger(&ZapLogger{logger: zapLogger}),
)
```

## Validation

Implement the `Validator` interface for request validation:

```go
type Validator interface {
    Validate(v any) error
}
```

```go
type MyValidator struct{}

func (v *MyValidator) Validate(data any) error {
    // Validation logic
    return nil
}

server := http.New(
    http.WithValidator(&MyValidator{}),
)
```

## Graceful Shutdown

The server handles graceful shutdown automatically:

1. Listens for SIGINT/SIGTERM signals
2. Executes `BeforeShutdown` hooks
3. Stops accepting new connections
4. Waits for active connections to finish (with timeout)
5. Executes `AfterShutdown` hooks

Shutdown timeout is configurable via `ServerConfig.GetShutdownTimeout()`.

## Plugin System

Plugins can extend server functionality:

```go
type MyPlugin struct{}

func (p *MyPlugin) Name() string {
    return "my-plugin"
}

func (p *MyPlugin) Install(server plugin.ServerInterface) error {
    // Install plugin logic
    return nil
}

// Note: Plugin system is designed for external use
// Plugins should be installed before calling server.Run()
```

## Best Practices

1. **Always use graceful shutdown**: The `Run()` method handles this automatically
2. **Use middleware for cross-cutting concerns**: Authentication, logging, etc.
3. **Implement custom error handlers**: For consistent error responses
4. **Use route groups**: For organizing related routes
5. **Set appropriate timeouts**: Prevent resource exhaustion
6. **Use health checks**: For Kubernetes/Docker deployments
7. **Implement lifecycle hooks**: For proper resource management

## Architecture

The wrapper is designed with the following principles:

- **Framework Agnostic**: All dependencies through interfaces
- **Extensible**: Plugin-based architecture
- **SOLID Principles**: Especially Open/Closed and Dependency Inversion
- **Clean Architecture**: Separation of concerns
- **Type Safe**: Full type safety with Go

## License

This package is part of the point project.

