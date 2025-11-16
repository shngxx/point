# Point Control Application

Приложение для управления точкой в реальном времени через WebSocket.

## Установка зависимостей

```bash
npm install
```

## Разработка

Запустите фронтенд в режиме разработки:

```bash
npm run dev
```

Фронтенд будет доступен на `http://localhost:3000` с проксированием запросов к бэкенду на `http://localhost:8080`.

## Сборка для продакшена

```bash
npm run build
```

Собранные файлы будут находиться в папке `web/`.

## Запуск бэкенда

```bash
go run cmd/app/main.go
```

Бэкенд будет доступен на `http://localhost:8080`.

## Архитектура

Приложение построено с использованием чистой архитектуры и DI (Dependency Injection).

### DI Container

Проект использует собственный DI контейнер:

```go
container := di.NewContainer()

// Регистрируем все конструкторы одним вызовом (порядок не важен!)
container.Provide(
    db.NewPointRepository,
    usecase.NewGetPointUC,
    usecase.NewMovePointUC,
    ws.NewHandler,
)

// Получаем нужный сервис - DI контейнер автоматически создаст все зависимости
handler := di.MustResolve[*ws.Handler](container)
```

**Основные возможности:**
- ✅ Автоматическое разрешение зависимостей
- ✅ Variadic конструкторы (множество конструкторов в одном вызове)
- ✅ Поддержка error в конструкторах
- ✅ Конструкторы могут возвращать несколько объектов
- ✅ Singleton по умолчанию
- ✅ Порядок регистрации не важен
- ✅ Thread-safe

