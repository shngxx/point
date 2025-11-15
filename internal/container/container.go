package container

import (
	"github.com/shngxx/point/internal/infrastructure/db"
	"github.com/shngxx/point/internal/usecase"
	"github.com/shngxx/point/internal/ws"
	"github.com/shngxx/point/pkg/di"
)

// SetupContainer настраивает DI контейнер и регистрирует все зависимости
func SetupContainer(container *di.Container) {
	container.Provide(db.NewPointRepository)
	container.Provide(usecase.NewGetPointUC)
	container.Provide(usecase.NewMovePointUC)
	container.Provide(ws.NewHandler)
}
