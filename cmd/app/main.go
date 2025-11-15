package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/shngxx/point/internal/container"
	httphandler "github.com/shngxx/point/internal/http"
	"github.com/shngxx/point/internal/usecase"
	"github.com/shngxx/point/internal/ws"
	"github.com/shngxx/point/pkg/di"
)

var c *di.Container

func main() {
	addr := flag.String("addr", ":8080", "Адрес сервера")
	flag.Parse()

	// Настройка контейнера
	c = di.NewContainer()
	container.SetupContainer(c)

	// Настройка роутера
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Регистрация маршрутов
	r.Get("/ws", di.MustResolveType[*ws.Handler](c).HandleWebSocket)
	r.Get("/api/point/{id}", httphandler.NewGetPointHandler(di.MustResolveType[*usecase.GetPointUC](c)))

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    *addr,
		Handler: r,
	}

	// Канал для получения сигналов ОС
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Printf("Сервер запущен на %s", *addr)
		log.Printf("Веб-интерфейс доступен по адресу http://localhost%s", *addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	// Ожидаем сигнал для graceful shutdown
	sig := <-sigChan
	log.Printf("Получен сигнал: %v. Начинаем graceful shutdown...", sig)

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Останавливаем сервер
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Ошибка при остановке сервера: %v", err)
	} else {
		log.Println("Сервер успешно остановлен")
	}
}
