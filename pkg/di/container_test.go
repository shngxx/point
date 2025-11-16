package di_test

import (
	"errors"
	"testing"

	"github.com/shngxx/point/pkg/di"
)

// Example 1: Simple constructor without dependencies
func TestProvide_SimpleConstructor(t *testing.T) {
	type Service struct {
		Name string
	}

	container := di.NewContainer()

	// Register constructor
	container.Provide(func() *Service {
		return &Service{Name: "test"}
	})

	// Resolve service
	service := di.MustResolve[*Service](container)
	if service.Name != "test" {
		t.Errorf("Expected Name='test', got '%s'", service.Name)
	}
}

// Example 2: Constructor with dependencies
func TestProvide_WithDependencies(t *testing.T) {
	type Database struct {
		Connected bool
	}

	type Repository struct {
		DB *Database
	}

	type Service struct {
		Repo *Repository
	}

	container := di.NewContainer()

	// Register all constructors in one call (order doesn't matter!)
	container.Provide(
		func(repo *Repository) *Service {
			return &Service{Repo: repo}
		},
		func(db *Database) *Repository {
			return &Repository{DB: db}
		},
		func() *Database {
			return &Database{Connected: true}
		},
	)

	// Resolve service - DI container will automatically create all dependencies
	service := di.MustResolve[*Service](container)
	if service.Repo == nil {
		t.Fatal("Repository was not created")
	}
	if service.Repo.DB == nil {
		t.Fatal("Database was not created")
	}
	if !service.Repo.DB.Connected {
		t.Error("Database should be connected")
	}
}

// Example 3: Constructor with error
func TestProvide_WithError(t *testing.T) {
	type Config struct {
		Valid bool
	}

	type Service struct {
		Config *Config
	}

	container := di.NewContainer()

	// Constructor can return error
	container.Provide(
		func() (*Config, error) {
			return &Config{Valid: true}, nil
		},
		func(cfg *Config) (*Service, error) {
			if !cfg.Valid {
				return nil, errors.New("invalid config")
			}
			return &Service{Config: cfg}, nil
		},
	)

	// Resolve service
	service := di.MustResolve[*Service](container)
	if service.Config == nil || !service.Config.Valid {
		t.Error("Config should be valid")
	}
}

// Example 4: Constructor with error that should panic
func TestProvide_WithErrorPanic(t *testing.T) {
	type Config struct{}
	type Service struct{}

	container := di.NewContainer()

	// Register constructor that returns an error
	container.Provide(
		func() (*Config, error) {
			return nil, errors.New("failed to create config")
		},
		func(cfg *Config) *Service {
			return &Service{}
		},
	)

	// Resolve attempt should panic due to constructor error
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic on constructor error")
		}
	}()

	di.MustResolve[*Service](container)
}

// Example 5: Constructor returns multiple values
func TestProvide_MultipleReturns(t *testing.T) {
	type Logger struct {
		Name string
	}

	type Database struct {
		Name string
	}

	type Service struct {
		Logger *Logger
		DB     *Database
	}

	container := di.NewContainer()

	// Constructor can return multiple objects
	container.Provide(
		func() (*Logger, *Database) {
			return &Logger{Name: "logger"}, &Database{Name: "db"}
		},
		func(logger *Logger, db *Database) *Service {
			return &Service{Logger: logger, DB: db}
		},
	)

	// Can resolve any of the returned types
	logger := di.MustResolve[*Logger](container)
	if logger.Name != "logger" {
		t.Errorf("Expected Name='logger', got '%s'", logger.Name)
	}

	db := di.MustResolve[*Database](container)
	if db.Name != "db" {
		t.Errorf("Expected Name='db', got '%s'", db.Name)
	}

	service := di.MustResolve[*Service](container)
	if service.Logger.Name != "logger" || service.DB.Name != "db" {
		t.Error("Service has incorrect dependencies")
	}
}

// Example 6: Constructor returns multiple values and error
func TestProvide_MultipleReturnsWithError(t *testing.T) {
	type Logger struct {
		Name string
	}

	type Database struct {
		Name string
	}

	container := di.NewContainer()

	// Constructor returns multiple objects and error
	container.Provide(
		func() (*Logger, *Database, error) {
			return &Logger{Name: "logger"}, &Database{Name: "db"}, nil
		},
	)

	logger := di.MustResolve[*Logger](container)
	if logger.Name != "logger" {
		t.Errorf("Expected Name='logger', got '%s'", logger.Name)
	}

	db := di.MustResolve[*Database](container)
	if db.Name != "db" {
		t.Errorf("Expected Name='db', got '%s'", db.Name)
	}
}

// Example 7: Singleton behavior (all dependencies are created once)
func TestProvide_SingletonBehavior(t *testing.T) {
	type Counter struct {
		Value int
	}

	callCount := 0
	container := di.NewContainer()

	container.Provide(func() *Counter {
		callCount++
		return &Counter{Value: callCount}
	})

	// Resolve multiple times
	counter1 := di.MustResolve[*Counter](container)
	counter2 := di.MustResolve[*Counter](container)

	// Constructor should be called only once (singleton)
	if callCount != 1 {
		t.Errorf("Constructor called %d times, expected 1", callCount)
	}

	// Both instances should be the same object
	if counter1 != counter2 {
		t.Error("Instances should be the same (singleton)")
	}

	if counter1.Value != 1 {
		t.Errorf("Expected Value=1, got %d", counter1.Value)
	}
}
