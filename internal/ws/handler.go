package ws

import (
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/shngxx/point/internal/domain/point"
	"github.com/shngxx/point/internal/usecase"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Разрешаем подключения с любого origin для разработки
	},
}

// GetPointService определяет интерфейс для получения информации о точке
// Интерфейс определен по месту потребления (в handler)
type GetPointService interface {
	Execute(id int) *usecase.PointInfo
}

// MovePointService определяет интерфейс для работы с движением точки
// Интерфейс определен по месту потребления (в handler)
type MovePointService interface {
	// Init запускает горутину для обработки движения точки
	Init(ctx context.Context, id int) <-chan *point.Point

	// Push добавляет команду перемещения в канал
	Push(cmd usecase.MoveCommand)
}

// Message представляет сообщение от клиента
type Message struct {
	X      int    `json:"x,omitempty"`
	Y      int    `json:"y,omitempty"`
	Action string `json:"action,omitempty"`
	DX     int    `json:"dx,omitempty"`
	DY     int    `json:"dy,omitempty"`
}

// PositionMessage представляет сообщение с позицией для клиента
type PositionMessage struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Handler обрабатывает WebSocket соединения
type Handler struct {
	getPointService  GetPointService
	movePointService MovePointService
	clients          map[*websocket.Conn]bool
	clientsMu        sync.RWMutex
}

// NewHandler создает новый WebSocket обработчик
func NewHandler(getPointService GetPointService, movePointService MovePointService) *Handler {
	return &Handler{
		getPointService:  getPointService,
		movePointService: movePointService,
		clients:          make(map[*websocket.Conn]bool),
	}
}

// HandleWebSocket обрабатывает WebSocket соединение
func (h *Handler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Ошибка при обновлении соединения: %v", err)
		return
	}

	// Регистрируем клиента
	h.clientsMu.Lock()
	h.clients[conn] = true
	h.clientsMu.Unlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer func() {
		cancel()
		h.clientsMu.Lock()
		delete(h.clients, conn)
		h.clientsMu.Unlock()
		conn.Close()
	}()

	log.Println("Новое WebSocket соединение установлено")

	// ID точки (по умолчанию 1)
	pointID := 1

	// Инициализируем обработку движения точки
	positionChan := h.movePointService.Init(ctx, pointID)

	// Читаем сообщения от клиента
	go func() {
		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("Ошибка чтения: %v", err)
				}
				cancel()
				break
			}

			// Если есть команда на движение, добавляем в канал
			if msg.DX != 0 || msg.DY != 0 {
				h.movePointService.Push(usecase.MoveCommand{
					ID: pointID,
					DX: msg.DX,
					DY: msg.DY,
				})
			}
		}
	}()

	// Отправляем обновления позиции клиенту
	for {
		select {
		case <-ctx.Done():
			return
		case pos := <-positionChan:
			h.sendPosition(conn, pos)
		}
	}
}

// BroadcastPosition отправляет позицию всем подключенным клиентам
// Используется для управления точкой с бэкенда
func (h *Handler) BroadcastPosition(pointID int) {
	pointInfo := h.getPointService.Execute(pointID)
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for conn := range h.clients {
		h.sendPosition(conn, pointInfo.Point)
	}
}

func (h *Handler) sendPosition(conn *websocket.Conn, pos *point.Point) {
	msg := PositionMessage{
		X: pos.X,
		Y: pos.Y,
	}
	if err := conn.WriteJSON(msg); err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
}
