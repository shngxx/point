package main

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	httphandler "github.com/shngxx/point/internal/http"
	"github.com/shngxx/point/internal/infrastructure/db"
	"github.com/shngxx/point/internal/usecase"
	"github.com/shngxx/point/internal/ws"
	"github.com/shngxx/point/pkg/config"
	"github.com/shngxx/point/pkg/di"
	"github.com/shngxx/point/pkg/http"
	httphooks "github.com/shngxx/point/pkg/http/hooks"
	logging "github.com/shngxx/point/pkg/log"
	wsmanager "github.com/shngxx/point/pkg/ws"
)

func main() {
	var cfg AppConfig
	config.LoadDefault(&cfg)

	// Setup DI container
	c := di.NewContainer()
	c.Provide(
		logging.New,
		wsmanager.NewManagerWithDefaults,
		http.NewWithDefaults,
		db.NewPointRepository,
		usecase.NewGetPointUC,
		usecase.NewMovePointUC,
		ws.NewHandler,
		httphandler.NewGetPointHandler,
	)

	// Register dependencies for server
	c.Supply(
		cfg.Server,
		cfg.Logger,
		usecase.MovePointConfig{
			BatchInterval: cfg.Point.BatchIntervalDuration(),
			SaveInterval:  cfg.Point.SaveIntervalDuration(),
		},
	)

	// Get dependencies from DI
	server := di.MustResolve[*http.Server](c)
	wsManager := di.MustResolve[*wsmanager.Manager](c)

	// Register all routes in a centralized location (routes.go)
	// Routes resolve their handlers from DI container automatically
	registerRoutes(server, c)

	// Register shutdown hook for WebSocket manager
	server.AddHook(httphooks.BeforeShutdown, func() error {
		return wsManager.Shutdown()
	})

	// Start server
	server.Start()
}

func registerRoutes(server *http.Server, c *di.Container) {
	// ============================================================================
	// WebSocket Routes
	// ============================================================================
	wsHandler := di.MustResolve[*ws.Handler](c)
	server.App().Get("/ws", websocket.New(wsHandler.Manager().HandleConnection))

	// ============================================================================
	// Point API Routes
	// ============================================================================
	getPointHandler := di.MustResolve[fiber.Handler](c)
	server.GET("/api/point/:id", getPointHandler)
	server.GET("/api/point", getPointHandler) // For case when id is not specified

}
