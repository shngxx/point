# Config Package

Universal package for loading configuration from YAML files with support for override via environment variables. Designed for use in Go microservices.

## Features

- ✅ Load configuration from YAML files
- ✅ Override values via environment variables
- ✅ Support for nested structures
- ✅ Load specific sections from a common config
- ✅ Use prefixes for environment variables
- ✅ Integration with DI container via `Supply`

## Installation

```bash
go get github.com/knadh/koanf/v2
go get github.com/knadh/koanf/parsers/yaml
go get github.com/knadh/koanf/providers/file
go get github.com/knadh/koanf/providers/env
```

## Quick Start

### 1. Basic Configuration Loading

```go
package main

import (
    "log"
    "github.com/shngxx/point/pkg/config"
)

type AppConfig struct {
    Host string `koanf:"host"`
    Port int    `koanf:"port"`
}

func main() {
    var cfg AppConfig
    if err := config.Load("config.yaml", &cfg); err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Host: %s, Port: %d", cfg.Host, cfg.Port)
}
```

**config.yaml:**
```yaml
host: localhost
port: 8080
```

### 2. Override via Environment Variables

```go
var cfg AppConfig
// Environment variables: APP_HOST, APP_PORT
err := config.LoadWithPrefix("config.yaml", &cfg, "APP_")
```

**Usage example:**
```bash
export APP_HOST=production.example.com
export APP_PORT=9090
./myapp
```

**Using default path (next to executable):**
```go
var cfg AppConfig
// Automatically loads from config.yaml next to the binary
err := config.LoadWithPrefixDefault(&cfg, "APP_")
```

### 3. Nested Structures

```go
type DatabaseConfig struct {
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    User     string `koanf:"user"`
    Password string `koanf:"password"`
}

type AppConfig struct {
    Database DatabaseConfig `koanf:"database"`
}
```

**config.yaml:**
```yaml
database:
  host: localhost
  port: 5432
  user: admin
  password: secret
```

**Environment variables:**
```bash
export APP_DATABASE_HOST=prod-db.example.com
export APP_DATABASE_PORT=3306
```

### 4. Loading a Specific Section

Useful when one YAML file contains configurations for multiple services:

```go
type DBConfig struct {
    Host string `koanf:"host"`
    Port int    `koanf:"port"`
}

var dbCfg DBConfig
// Load only "database" section
// Environment variables: DB_HOST, DB_PORT
err := config.LoadSection("config.yaml", "database", &dbCfg, "DB_")
```

**config.yaml:**
```yaml
database:
  host: localhost
  port: 5432
  
redis:
  host: localhost
  port: 6379
```

## Integration with DI Container

### Using with Supply

```go
package main

import (
    "github.com/shngxx/point/pkg/config"
    "github.com/shngxx/point/pkg/di"
)

type ServerConfig struct {
    Host string `koanf:"host"`
    Port int    `koanf:"port"`
}

type AppConfig struct {
    Server ServerConfig `koanf:"server"`
}

func main() {
    // 1. Load configuration
    var cfg AppConfig
    config.LoadWithPrefix("config.yaml", &cfg, "APP_")
    
    // 2. Create DI container
    container := di.NewContainer()
    
    // 3. Supply: Register ready configuration
    container.Supply(cfg.Server)
    
    // 4. Provide: Register services that use configuration
    container.Provide(NewHTTPServer)
}

// Constructor automatically receives configuration via DI
func NewHTTPServer(cfg ServerConfig) *http.Server {
    addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
    return &http.Server{Addr: addr}
}
```

### Supply vs Provide

| Supply | Provide |
|--------|---------|
| Accepts ready values | Accepts constructor functions |
| For configuration, constants | For services with dependencies |
| `container.Supply(config)` | `container.Provide(NewService)` |
| Values available immediately | Lazy initialization |

## Complete Application Example

**config.yaml:**
```yaml
server:
  host: localhost
  port: 8080
  readTimeout: 30
  writeTimeout: 30

database:
  host: localhost
  port: 5432
  name: mydb
  user: user
  password: pass

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

**main.go:**
```go
package main

import (
    "fmt"
    "log"
    "github.com/shngxx/point/pkg/config"
    "github.com/shngxx/point/pkg/di"
)

// Define configuration structures
type ServerConfig struct {
    Host         string `koanf:"host"`
    Port         int    `koanf:"port"`
    ReadTimeout  int    `koanf:"readTimeout"`
    WriteTimeout int    `koanf:"writeTimeout"`
}

type DBConfig struct {
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    Name     string `koanf:"name"`
    User     string `koanf:"user"`
    Password string `koanf:"password"`
}

type RedisConfig struct {
    Host     string `koanf:"host"`
    Port     int    `koanf:"port"`
    Password string `koanf:"password"`
    DB       int    `koanf:"db"`
}

type AppConfig struct {
    Server ServerConfig `koanf:"server"`
    DB     DBConfig     `koanf:"database"`
    Redis  RedisConfig  `koanf:"redis"`
}

func main() {
    // Load configuration
    var cfg AppConfig
    if err := config.LoadWithPrefix("config.yaml", &cfg, "APP_"); err != nil {
        log.Fatalf("Error loading configuration: %v", err)
    }

    // Setup DI container
    container := di.NewContainer()
    
    // Supply: register configurations
    container.Supply(cfg.Server, cfg.DB, cfg.Redis)
    
    // Provide: register services
    container.Provide(
        NewDatabase,
        NewRedisClient,
        NewHTTPServer,
    )
    
    // Get server and start
    server := di.MustResolve[*HTTPServer](container)
    server.Start()
}

// Constructors automatically receive configuration
func NewDatabase(cfg DBConfig) (*Database, error) {
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
        cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
    // ... connect to database
    return &Database{}, nil
}

func NewRedisClient(cfg RedisConfig) (*RedisClient, error) {
    addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
    // ... connect to Redis
    return &RedisClient{}, nil
}

func NewHTTPServer(cfg ServerConfig, db *Database, redis *RedisClient) *HTTPServer {
    return &HTTPServer{
        addr: fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
        db:   db,
        redis: redis,
    }
}
```

## Environment Variables

### Name Format

Environment variables are formed according to the following rules:

1. Add prefix (if specified): `APP_`
2. Nested keys are separated by `_`: `SERVER_HOST`
3. Final format: `APP_SERVER_HOST`

### Examples

**Configuration:**
```yaml
server:
  host: localhost
  port: 8080
database:
  connection:
    timeout: 30
```

**Environment variables (with APP_ prefix):**
```bash
APP_SERVER_HOST=production.example.com
APP_SERVER_PORT=9090
APP_DATABASE_CONNECTION_TIMEOUT=60
```

## Override Hierarchy

1. **Base values**: YAML file
2. **Overrides**: Environment variables

Environment variables always take precedence over values from the YAML file.

## Best Practices

### 1. Use separate types for each service

```go
type DBConfig struct { ... }
type RedisConfig struct { ... }
type ServerConfig struct { ... }
```

Each configuration is a separate type for type-safety in DI.

### 2. Prefixes for environment variables

Always use prefixes to avoid conflicts:
```go
config.LoadWithPrefix("config.yaml", &cfg, "MYAPP_")
```

### 3. Configuration validation

```go
if err := config.Load("config.yaml", &cfg); err != nil {
    log.Fatal(err)
}
if err := validateConfig(cfg); err != nil {
    log.Fatal(err)
}
```

### 4. Sensitive data

Store sensitive data (passwords, tokens) in environment variables:

```yaml
database:
  host: localhost
  port: 5432
  # password is passed via DATABASE_PASSWORD
```

### 5. Different configs for different environments

```go
configFile := "config.yaml"
if env := os.Getenv("ENV"); env != "" {
    configFile = fmt.Sprintf("config.%s.yaml", env)
}
config.Load(configFile, &cfg)
```

## Usage Examples in Different Scenarios

### Microservice with one database

```go
type Config struct {
    Server ServerConfig `koanf:"server"`
    DB     DBConfig     `koanf:"database"`
}

var cfg Config
config.Load("config.yaml", &cfg)
container.Supply(cfg.Server, cfg.DB)
```

### Microservice with multiple storages

```go
type Config struct {
    Server ServerConfig `koanf:"server"`
    DB     DBConfig     `koanf:"database"`
    Redis  RedisConfig  `koanf:"redis"`
    Cache  CacheConfig  `koanf:"cache"`
}

var cfg Config
config.Load("config.yaml", &cfg)
container.Supply(cfg.Server, cfg.DB, cfg.Redis, cfg.Cache)
```

### Loading sections from common config

```go
// Common config stores settings for all services
var dbCfg DBConfig
config.LoadSection("common-config.yaml", "database", &dbCfg, "DB_")

var redisCfg RedisConfig
config.LoadSection("common-config.yaml", "redis", &redisCfg, "REDIS_")
```

## API Reference

### Load

```go
func Load(configPath string, target any) error
```

Loads configuration from a YAML file with override via environment variables.

### LoadWithPrefix

```go
func LoadWithPrefix(configPath string, target any, envPrefix string) error
```

Loads configuration with the specified prefix for environment variables.

### LoadSection

```go
func LoadSection(configPath string, section string, target any, envPrefix string) error
```

Loads a specific section from a YAML file.

### LoadDefault

```go
func LoadDefault(target any)
```

Loads configuration from the default `config.yaml` file located next to the executable binary. This is a convenience function that uses the default path, so you don't need to specify the config file path. Panics if configuration cannot be loaded.

**Example:**
```go
var cfg AppConfig
config.LoadDefault(&cfg)
```

### LoadWithPrefixDefault

```go
func LoadWithPrefixDefault(target any, envPrefix string)
```

Loads configuration from the default `config.yaml` file (next to the executable) with override via environment variables using the specified prefix. Panics if configuration cannot be loaded.

**Example:**
```go
var cfg AppConfig
// Override via: APP_SERVER_HOST, APP_SERVER_PORT
config.LoadWithPrefixDefault(&cfg, "APP_")
```

### LoadSectionDefault

```go
func LoadSectionDefault(section string, target any, envPrefix string) error
```

Loads a specific section from the default `config.yaml` file (next to the executable) with override via environment variables.

**Example:**
```go
var dbCfg DBConfig
// Override via: DB_HOST, DB_PORT
err := config.LoadSectionDefault("database", &dbCfg, "DB_")
```

**Note:** The `*Default` functions automatically locate `config.yaml` in the same directory as the executable binary. If the executable path cannot be determined, they fall back to `config.yaml` in the current working directory.

## Troubleshooting

### Values are not overridden from env

Check:
1. Correct variable name format (PREFIX_KEY_SUBKEY)
2. That the variable is exported: `export VAR_NAME=value`
3. That the correct prefix is used

### Deserialization error

Check:
1. `koanf:"..."` tags on all structure fields
2. Type matching (int, string, bool)
3. YAML syntax validity

### Type conflicts in DI

Make sure each configuration has a unique type:
```go
type DBConfig struct { ... }
type RedisConfig struct { ... }
// DO NOT use the same type twice
```

