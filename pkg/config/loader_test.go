package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestLoad tests basic configuration loading from YAML
func TestLoad(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
host: localhost
port: 8080
debug: true
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Define configuration structure
	type Config struct {
		Host  string `koanf:"host"`
		Port  int    `koanf:"port"`
		Debug bool   `koanf:"debug"`
	}

	// Load configuration
	var cfg Config
	if err := Load(configPath, &cfg); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check values
	if cfg.Host != "localhost" {
		t.Errorf("Host = %v, expected localhost", cfg.Host)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port = %v, expected 8080", cfg.Port)
	}
	if !cfg.Debug {
		t.Errorf("Debug = %v, expected true", cfg.Debug)
	}
}

// TestLoadWithPrefix tests loading with override via environment variables
func TestLoadWithPrefix(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
host: localhost
port: 8080
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set environment variables for override with unique prefix
	os.Setenv("TEST_APP_CFG_HOST", "production.example.com")
	os.Setenv("TEST_APP_CFG_PORT", "9090")
	defer func() {
		os.Unsetenv("TEST_APP_CFG_HOST")
		os.Unsetenv("TEST_APP_CFG_PORT")
	}()

	// Define configuration structure
	type Config struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	// Load configuration with prefix
	var cfg Config
	if err := LoadWithPrefix(configPath, &cfg, "TEST_APP_CFG_"); err != nil {
		t.Fatalf("LoadWithPrefix() error = %v", err)
	}

	// Check that values are overridden from environment variables
	if cfg.Host != "production.example.com" {
		t.Errorf("Host = %v, expected production.example.com (from env)", cfg.Host)
	}
	if cfg.Port != 9090 {
		t.Errorf("Port = %v, expected 9090 (from env)", cfg.Port)
	}
}

// TestLoadWithNestedStructure tests loading nested structures
func TestLoadWithNestedStructure(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
database:
  host: db.local
  port: 5432
  credentials:
    user: admin
    password: secret
server:
  host: 0.0.0.0
  port: 8080
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Define configuration structure
	type Credentials struct {
		User     string `koanf:"user"`
		Password string `koanf:"password"`
	}

	type DatabaseConfig struct {
		Host        string      `koanf:"host"`
		Port        int         `koanf:"port"`
		Credentials Credentials `koanf:"credentials"`
	}

	type ServerConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	type Config struct {
		Database DatabaseConfig `koanf:"database"`
		Server   ServerConfig   `koanf:"server"`
	}

	// Load configuration
	var cfg Config
	if err := Load(configPath, &cfg); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check values
	if cfg.Database.Host != "db.local" {
		t.Errorf("Database.Host = %v, expected db.local", cfg.Database.Host)
	}
	if cfg.Database.Port != 5432 {
		t.Errorf("Database.Port = %v, expected 5432", cfg.Database.Port)
	}
	if cfg.Database.Credentials.User != "admin" {
		t.Errorf("Database.Credentials.User = %v, expected admin", cfg.Database.Credentials.User)
	}
	if cfg.Database.Credentials.Password != "secret" {
		t.Errorf("Database.Credentials.Password = %v, expected secret", cfg.Database.Credentials.Password)
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("Server.Host = %v, expected 0.0.0.0", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("Server.Port = %v, expected 8080", cfg.Server.Port)
	}
}

// TestLoadSection tests loading a specific section
func TestLoadSection(t *testing.T) {
	// Create temporary YAML file with multiple sections
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
database:
  host: db.local
  port: 5432
redis:
  host: redis.local
  port: 6379
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Define structure for database section
	type DBConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	// Load only database section
	var dbCfg DBConfig
	if err := LoadSection(configPath, "database", &dbCfg, "TEST_DB_CFG_"); err != nil {
		t.Fatalf("LoadSection() error = %v", err)
	}

	// Check values
	if dbCfg.Host != "db.local" {
		t.Errorf("Host = %v, expected db.local", dbCfg.Host)
	}
	if dbCfg.Port != 5432 {
		t.Errorf("Port = %v, expected 5432", dbCfg.Port)
	}

	// Define structure for redis section
	type RedisConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	// Load only redis section
	var redisCfg RedisConfig
	if err := LoadSection(configPath, "redis", &redisCfg, "TEST_REDIS_CFG_"); err != nil {
		t.Fatalf("LoadSection() error = %v", err)
	}

	// Check values
	if redisCfg.Host != "redis.local" {
		t.Errorf("Host = %v, expected redis.local", redisCfg.Host)
	}
	if redisCfg.Port != 6379 {
		t.Errorf("Port = %v, expected 6379", redisCfg.Port)
	}
}

// TestLoadSectionWithEnvOverride tests section override via environment variables
func TestLoadSectionWithEnvOverride(t *testing.T) {
	// Create temporary YAML file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	yamlContent := `
database:
  host: localhost
  port: 5432
`
	if err := os.WriteFile(configPath, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Set environment variables with unique prefix for test
	os.Setenv("TEST_SECTION_DB_HOST", "prod-db.example.com")
	os.Setenv("TEST_SECTION_DB_PORT", "3306")
	defer func() {
		os.Unsetenv("TEST_SECTION_DB_HOST")
		os.Unsetenv("TEST_SECTION_DB_PORT")
	}()

	// Define configuration structure
	type DBConfig struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	// Load section with prefix for env variables
	var cfg DBConfig
	if err := LoadSection(configPath, "database", &cfg, "TEST_SECTION_DB_"); err != nil {
		t.Fatalf("LoadSection() error = %v", err)
	}

	// Check that values are overridden
	if cfg.Host != "prod-db.example.com" {
		t.Errorf("Host = %v, expected prod-db.example.com (from env)", cfg.Host)
	}
	if cfg.Port != 3306 {
		t.Errorf("Port = %v, expected 3306 (from env)", cfg.Port)
	}
}

// TestLoadNonExistentFile tests error handling when loading non-existent file
func TestLoadNonExistentFile(t *testing.T) {
	type Config struct {
		Host string `koanf:"host"`
	}

	var cfg Config
	err := Load("/non/existent/path.yaml", &cfg)
	if err == nil {
		t.Error("Load() should return error for non-existent file")
	}
}

// TestLoadInvalidYAML tests handling of invalid YAML
func TestLoadInvalidYAML(t *testing.T) {
	// Create temporary file with invalid YAML
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `
host: localhost
  port: 8080
invalid yaml structure
`
	if err := os.WriteFile(configPath, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	type Config struct {
		Host string `koanf:"host"`
		Port int    `koanf:"port"`
	}

	var cfg Config
	err := Load(configPath, &cfg)
	if err == nil {
		t.Error("Load() should return error for invalid YAML")
	}
}
