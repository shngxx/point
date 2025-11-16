# WebSocket SDK Wrapper

A production-ready, extensible WebSocket wrapper over [gofiber/websocket](https://github.com/gofiber/websocket) for Go applications, following the same patterns as `pkg/http`.

## Features

- 🚀 **Simple API**: Clean and intuitive interface
- 🔌 **Extensible**: Middleware and hooks support
- 🛡️ **Production Ready**: Graceful shutdown, connection management, error handling
- 🔒 **Framework Agnostic**: All dependencies through interfaces
- 📦 **Room/Channel Pattern**: Group connections for targeted broadcasting
- 🎯 **Type Safe**: Full type safety with Go
- 🔄 **Message Routing**: Action-based message handling

## Quick Start

### Minimal Example

```go
package main

import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
    "github.com/shngxx/point/pkg/ws"
    "github.com/shngxx/point/pkg/ws/middleware"
    "github.com/shngxx/point/pkg/log"
)

func main() {
    logger := log.New(...)
    
    // Create WebSocket manager
    wsManager := ws.NewManager(
        ws.WithLogger(logger),
        ws.WithMiddleware(
            middleware.Logger(logger),
            middleware.Recovery(logger),
        ),
    )
    
    // Register message handlers
    wsManager.HandleMessage("subscribe", func(conn *ws.Connection, msg *ws.Message) error {
        // Handle subscribe action
        return nil
    })
    
    // Create HTTP server
    app := fiber.New()
    
    // Register WebSocket endpoint
    app.Get("/ws", websocket.New(wsManager.HandleConnection))
    
    app.Listen(":8080")
}
```

### Advanced Example

```go
package main

import (
    "context"
    "encoding/json"
    
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
    "github.com/shngxx/point/pkg/ws"
    "github.com/shngxx/point/pkg/ws/hooks"
    "github.com/shngxx/point/pkg/ws/middleware"
    "github.com/shngxx/point/pkg/log"
)

func main() {
    logger := log.New(...)
    
    // Custom configuration
    cfg := &ws.DefaultConfig{
        PingInterval:         60 * time.Second,
        PongTimeout:          10 * time.Second,
        ReadBufferSize:       4096,
        WriteBufferSize:      4096,
        MaxConnectionsPerRoom: 100,
        ShutdownTimeout:      30 * time.Second,
    }
    
    // Create WebSocket manager
    wsManager := ws.NewManager(
        ws.WithConfig(cfg),
        ws.WithLogger(logger),
        ws.WithMiddleware(
            middleware.Logger(logger),
            middleware.Recovery(logger),
        ),
        ws.WithHook(hooks.OnConnect, func(conn ws.ConnectionInterface, data ...any) error {
            logger.Info("Client connected")
            return nil
        }),
        ws.WithHook(hooks.OnDisconnect, func(conn ws.ConnectionInterface, data ...any) error {
            logger.Info("Client disconnected")
            return nil
        }),
    )
    
    // Register message handlers
    wsManager.HandleMessage("subscribe", handleSubscribe)
    wsManager.HandleMessage("unsubscribe", handleUnsubscribe)
    wsManager.HandleMessage("ping", handlePing)
    
    // Create HTTP server
    app := fiber.New()
    app.Get("/ws", websocket.New(wsManager.HandleConnection))
    
    // Graceful shutdown
    defer wsManager.Shutdown()
    
    app.Listen(":8080")
}

func handleSubscribe(conn *ws.Connection, msg *ws.Message) error {
    var data struct {
        RoomID string `json:"room_id"`
    }
    if err := json.Unmarshal(msg.Data, &data); err != nil {
        return err
    }
    
    // Get manager from connection metadata or pass it differently
    // For this example, assume manager is accessible
    // manager.JoinRoom(conn, data.RoomID)
    
    return conn.WriteJSON(map[string]string{
        "status": "subscribed",
        "room": data.RoomID,
    })
}
```

## Configuration

### Manager Configuration

The manager can be configured using the `ManagerConfig` interface or the `DefaultConfig` struct:

```go
type ManagerConfig interface {
    GetPingInterval() time.Duration
    GetPongTimeout() time.Duration
    GetReadBufferSize() int
    GetWriteBufferSize() int
    GetMaxConnectionsPerRoom() int
    GetShutdownTimeout() time.Duration
}
```

Example:

```go
cfg := &ws.DefaultConfig{
    PingInterval:         60 * time.Second,
    PongTimeout:          10 * time.Second,
    ReadBufferSize:       4096,
    WriteBufferSize:      4096,
    MaxConnectionsPerRoom: 100,
    ShutdownTimeout:      30 * time.Second,
}

wsManager := ws.NewManager(ws.WithConfig(cfg))
```

### Functional Options

The manager uses the functional options pattern for configuration:

- `WithLogger(logger Logger)` - Set custom logger
- `WithConfig(cfg ManagerConfig)` - Set manager configuration
- `WithMiddleware(mw ...middleware.Handler)` - Set global middleware
- `WithHook(hookType HookType, fn HookFunc)` - Register lifecycle hook

## Connection Management

### Connection Wrapper

Each WebSocket connection is wrapped in a `Connection` object that provides:

- **Metadata Storage**: Store custom data per connection
- **Subscription Tracking**: Track which rooms a connection is in
- **Typed Messages**: `ReadJSON()` and `WriteJSON()` methods
- **Context**: Cancellation support via context

```go
// Set connection metadata
conn.SetMetadata("user_id", "123")
conn.SetMetadata("session_id", "abc")

// Get metadata
userID, ok := conn.GetMetadata("user_id")

// Check subscriptions
rooms := conn.GetSubscriptions()
isSubscribed := conn.IsSubscribed("room_123")
```

## Room Management

### Rooms/Channels

Rooms allow you to group connections for targeted broadcasting:

```go
// Join a room
manager.JoinRoom(conn, "workflow_123")

// Leave a room
manager.LeaveRoom(conn, "workflow_123")

// Broadcast to a room
manager.BroadcastToRoom("workflow_123", event)

// Broadcast to all connections
manager.BroadcastToAll(event)

// Send to specific connection
manager.SendToConnection(conn, message)
```

### Room Use Cases

- **Workflow Execution**: One room per `workflow_execution_id`
- **Chat Rooms**: One room per chat channel
- **Game Sessions**: One room per game instance
- **Document Collaboration**: One room per document

## Message Routing

### Message Format

Messages should follow this structure:

```json
{
  "action": "subscribe",
  "data": {
    "room_id": "workflow_123"
  }
}
```

### Registering Handlers

```go
wsManager.HandleMessage("subscribe", func(conn *ws.Connection, msg *ws.Message) error {
    var data struct {
        RoomID string `json:"room_id"`
    }
    json.Unmarshal(msg.Data, &data)
    
    // Join room
    manager.JoinRoom(conn, data.RoomID)
    
    // Send confirmation
    return conn.WriteJSON(map[string]string{
        "status": "subscribed",
        "room": data.RoomID,
    })
})
```

### Message Handler Signature

```go
type MessageHandler func(conn *Connection, message *Message) error
```

## Middleware

### Built-in Middleware

#### Logger

Logs WebSocket connections:

```go
wsManager := ws.NewManager(
    ws.WithMiddleware(middleware.Logger(logger)),
)
```

#### Recovery

Recovers from panics:

```go
wsManager := ws.NewManager(
    ws.WithMiddleware(middleware.Recovery(logger)),
)
```

### Custom Middleware

Create custom middleware by implementing the `middleware.Handler` type:

```go
func MyMiddleware() middleware.Handler {
    return func(c middleware.ConnectionInterface) error {
        // Before connection processing
        // ...
        
        // Continue (middleware is called before connection handling)
        return nil
    }
}
```

## Lifecycle Hooks

Register hooks for connection lifecycle events:

```go
wsManager := ws.NewManager(
    ws.WithHook(hooks.OnConnect, func(conn ws.ConnectionInterface, data ...any) error {
        logger.Info("Client connected")
        return nil
    }),
    ws.WithHook(hooks.OnDisconnect, func(conn ws.ConnectionInterface, data ...any) error {
        logger.Info("Client disconnected")
        return nil
    }),
    ws.WithHook(hooks.OnJoinRoom, func(conn ws.ConnectionInterface, data ...any) error {
        roomID := data[0].(string)
        logger.Info("Client joined room", log.Field{Key: "room", Value: roomID})
        return nil
    }),
)
```

Available hook types:
- `hooks.OnConnect` - After connection established
- `hooks.OnDisconnect` - Before connection closed
- `hooks.OnMessage` - Before message processing
- `hooks.OnError` - When error occurs
- `hooks.OnJoinRoom` - When connection joins a room
- `hooks.OnLeaveRoom` - When connection leaves a room

## Broadcasting

### Broadcast Patterns

**Room-based Broadcasting** (recommended for most cases):
```go
// Broadcast to specific room
manager.BroadcastToRoom("workflow_123", event)
```

**Global Broadcasting**:
```go
// Broadcast to all connections
manager.BroadcastToAll(event)
```

**Unicast**:
```go
// Send to specific connection
manager.SendToConnection(conn, message)
```

### Event Structure

For workflow engine use case:

```go
type WorkflowEvent struct {
    Type                string      `json:"type"` // "node.started", "node.completed"
    WorkflowExecutionID string      `json:"workflow_execution_id"`
    NodeID              string      `json:"node_id"`
    Data                any `json:"data"`
    Timestamp           time.Time   `json:"timestamp"`
}

// From runner/usecase
event := WorkflowEvent{
    Type:                "node.started",
    WorkflowExecutionID: "exec_123",
    NodeID:              "node_456",
    Data:                map[string]any{"status": "running"},
    Timestamp:           time.Now(),
}

manager.BroadcastToRoom("workflow_exec_123", event)
```

## Integration with HTTP Server

### Fiber Integration

```go
import (
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/websocket/v2"
    "github.com/shngxx/point/pkg/ws"
)

func main() {
    app := fiber.New()
    wsManager := ws.NewManager(...)
    
    app.Get("/ws", websocket.New(wsManager.HandleConnection))
    
    // Graceful shutdown
    defer wsManager.Shutdown()
    
    app.Listen(":8080")
}
```

### With pkg/http Integration

```go
import (
    "github.com/shngxx/point/pkg/http"
    "github.com/shngxx/point/pkg/http/hooks"
    "github.com/shngxx/point/pkg/ws"
)

func main() {
    httpServer := http.New(...)
    wsManager := ws.NewManager(...)
    
    // Register WebSocket endpoint
    httpServer.App().Get("/ws", websocket.New(wsManager.HandleConnection))
    
    // Register shutdown hook
    httpServer.AddHook(hooks.BeforeShutdown, func() error {
        return wsManager.Shutdown()
    })
    
    httpServer.Start()
}
```

## Graceful Shutdown

The manager handles graceful shutdown:

1. Stops accepting new connections
2. Closes all existing connections
3. Waits for active operations to complete (with timeout)
4. Cleans up resources

```go
// Manual shutdown
wsManager.Shutdown()

// Or integrate with HTTP server lifecycle
httpServer.AddHook(hooks.BeforeShutdown, func() error {
    return wsManager.Shutdown()
})
```

## Best Practices

1. **Use Rooms for Grouping**: Group connections by business entity (workflow_id, chat_id, etc.)
2. **Register Message Handlers**: Use action-based routing for different message types
3. **Use Middleware**: For cross-cutting concerns (auth, logging, rate limiting)
4. **Implement Hooks**: For connection lifecycle management
5. **Store Metadata**: Use connection metadata for user context, session data
6. **Handle Errors**: Always handle errors in message handlers
7. **Graceful Shutdown**: Always call `Shutdown()` on application exit

## Architecture

The wrapper is designed with the following principles:

- **Framework Agnostic**: All dependencies through interfaces
- **Extensible**: Middleware and hooks for customization
- **SOLID Principles**: Especially Open/Closed and Dependency Inversion
- **Clean Architecture**: Separation of concerns
- **Type Safe**: Full type safety with Go
- **Room/Channel Pattern**: For scalable broadcasting

## Future Enhancements

- Redis Pub/Sub integration for horizontal scaling
- Kafka integration for event streaming
- Connection authentication middleware
- Rate limiting middleware
- Metrics and monitoring hooks

## License

This package is part of the point project.

