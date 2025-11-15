package usecase

import (
	"context"
	"log"
	"time"

	"github.com/shngxx/point/internal/domain/point"
)

// MoveCommand представляет команду на перемещение точки
type MoveCommand struct {
	ID int
	DX int
	DY int
}

// MovePointUC реализует сценарий использования: пошаговое движение точки
type MovePointUC struct {
	pointRepository point.PointRepository
	moveChan        chan MoveCommand
}

// NewMovePointUC создает новый usecase для пошагового движения точки
func NewMovePointUC(repository point.PointRepository) *MovePointUC {
	return &MovePointUC{
		pointRepository: repository,
		moveChan:        make(chan MoveCommand, 50), // Увеличен буфер для плавности
	}
}

// Init запускает горутину для обработки движения точки
// Вызывается один раз при активации WebSocket соединения
// Возвращает канал для получения обновлений позиции для этого конкретного соединения
func (u *MovePointUC) Init(ctx context.Context, id int) <-chan *point.Point {
	positionChan := make(chan *point.Point, 5)
	go u.processMoves(ctx, id, positionChan)
	return positionChan
}

// Push добавляет команду перемещения в канал
func (u *MovePointUC) Push(cmd MoveCommand) {
	select {
	case u.moveChan <- cmd:
	default:
		// Канал переполнен, игнорируем команду
	}
}

// processMoves обрабатывает команды перемещения в бесконечном цикле
// positionChan - канал для отправки обновлений позиции конкретному клиенту
func (u *MovePointUC) processMoves(ctx context.Context, id int, positionChan chan<- *point.Point) {
	ticker := time.NewTicker(5 * time.Second) // Сохраняем каждые 5 секунд
	defer ticker.Stop()
	defer close(positionChan)

	// Таймер для батчинга команд (обрабатываем накопленные команды каждые 16ms ~60 FPS)
	batchTicker := time.NewTicker(16 * time.Millisecond)
	defer batchTicker.Stop()

	var pendingCommands []MoveCommand
	lastSentPos := &point.Point{X: -1, Y: -1} // Для отслеживания изменений

	for {
		select {
		case <-ctx.Done():
			return
		case cmd := <-u.moveChan:
			// Накапливаем команды для батчинга
			pendingCommands = append(pendingCommands, cmd)
		case <-batchTicker.C:
			// Обрабатываем накопленные команды батчем
			if len(pendingCommands) > 0 {
				p := u.pointRepository.Get(id)

				// Применяем все команды последовательно
				for _, cmd := range pendingCommands {
					p.Move(cmd.DX, cmd.DY)
				}
				pendingCommands = pendingCommands[:0] // Очищаем слайс

				// Отправляем обновление только если позиция изменилась
				if p.X != lastSentPos.X || p.Y != lastSentPos.Y {
					lastSentPos.X = p.X
					lastSentPos.Y = p.Y

					select {
					case positionChan <- &point.Point{X: p.X, Y: p.Y}:
					default:
						// Канал переполнен, игнорируем
					}
				}
			}
		case <-ticker.C:
			// Периодически сохраняем положение точки
			p := u.pointRepository.Get(id)
			u.pointRepository.Save(id, p)
			log.Printf("id=%d: точка успешно сохранена: { x: %d, y: %d }", id, p.X, p.Y)
		}
	}
}
