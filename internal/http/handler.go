package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/shngxx/point/internal/usecase"
)

// GetPointService определяет интерфейс для получения информации о точке
type GetPointService interface {
	Execute(id int) *usecase.PointInfo
}

// NewGetPointHandler создает handler для получения информации о точке
func NewGetPointHandler(service GetPointService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if id == "" {
			id = "1" // По умолчанию ID = 1
		}

		var pointID int
		if _, err := fmt.Sscanf(id, "%d", &pointID); err != nil {
			http.Error(w, "Неверный ID точки", http.StatusBadRequest)
			return
		}

		pointInfo := service.Execute(pointID)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pointInfo)
	}
}
