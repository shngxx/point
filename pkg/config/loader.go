package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

// getDefaultConfigPath returns the path to config.yaml next to the executable.
// If the executable path cannot be determined, it falls back to "config.yaml" in the current directory.
func getDefaultConfigPath() string {
	execPath, err := os.Executable()
	if err != nil {
		// Fallback to current directory if executable path cannot be determined
		return "config.yaml"
	}
	execDir := filepath.Dir(execPath)
	return filepath.Join(execDir, "config.yaml")
}

// Load loads configuration from a YAML file with override via environment variables.
// Environment variables are automatically determined from the configuration structure.
//
// Parameters:
//   - configPath: path to the YAML configuration file
//   - target: pointer to the structure into which the configuration will be loaded
//
// Example:
//
//	type DBConfig struct {
//	    Host string `koanf:"host"`
//	    Port int    `koanf:"port"`
//	}
//
//	var cfg DBConfig
//	err := config.Load("config.yaml", &cfg)
func Load(configPath string, target any) error {
	return LoadWithPrefix(configPath, target, "")
}

// LoadWithPrefix loads configuration from a YAML file with override via environment variables,
// using the specified prefix for environment variables.
//
// Parameters:
//   - configPath: path to the YAML configuration file
//   - target: pointer to the structure into which the configuration will be loaded
//   - envPrefix: prefix for environment variables (e.g., "APP" for APP_HOST, APP_PORT)
//
// Override hierarchy (from lowest to highest):
//  1. YAML file
//  2. Environment variables
//
// Environment variables are formed as follows:
//   - Nested structures are separated by "_"
//   - If a prefix is specified, it is added at the beginning: PREFIX_KEY
//   - Example: for structure Server.Host with prefix "APP" -> APP_SERVER_HOST
//
// Example:
//
//	type Config struct {
//	    Server struct {
//	        Host string `koanf:"host"`
//	        Port int    `koanf:"port"`
//	    } `koanf:"server"`
//	}
//
//	var cfg Config
//	// Override via: APP_SERVER_HOST, APP_SERVER_PORT
//	err := config.LoadWithPrefix("config.yaml", &cfg, "APP")
func LoadWithPrefix(configPath string, target any, envPrefix string) error {
	k := koanf.New(".")

	// 1. Load configuration from YAML file
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading configuration from file %s: %w", configPath, err)
	}

	// 2. Override with values from environment variables
	// Variable format: PREFIX_KEY1_KEY2 (where . is replaced with _)
	// Callback function to transform environment variable names into configuration keys
	envCb := func(s string) string {
		// Remove prefix if present
		if envPrefix != "" && strings.HasPrefix(s, envPrefix) {
			s = strings.TrimPrefix(s, envPrefix)
		}
		// Transform SERVER_HOST -> server.host
		return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	}

	if err := k.Load(env.Provider("", ".", envCb), nil); err != nil {
		return fmt.Errorf("error loading environment variables: %w", err)
	}

	// 3. Unmarshal configuration into target structure
	if err := k.Unmarshal("", target); err != nil {
		return fmt.Errorf("error deserializing configuration: %w", err)
	}

	return nil
}

// LoadDefault loads configuration from the default config.yaml file (next to the executable)
// with override via environment variables.
// Environment variables are automatically determined from the configuration structure.
// Panics if configuration cannot be loaded.
//
// Parameters:
//   - target: pointer to the structure into which the configuration will be loaded
//
// Example:
//
//	type DBConfig struct {
//	    Host string `koanf:"host"`
//	    Port int    `koanf:"port"`
//	}
//
//	var cfg DBConfig
//	config.LoadDefault(&cfg)
func LoadDefault(target any) {
	LoadWithPrefixDefault(target, "")
}

// LoadWithPrefixDefault loads configuration from the default config.yaml file (next to the executable)
// with override via environment variables, using the specified prefix for environment variables.
// Panics if configuration cannot be loaded.
//
// Parameters:
//   - target: pointer to the structure into which the configuration will be loaded
//   - envPrefix: prefix for environment variables (e.g., "APP_" for APP_HOST, APP_PORT)
//
// Override hierarchy (from lowest to highest):
//  1. YAML file
//  2. Environment variables
//
// Environment variables are formed as follows:
//   - Nested structures are separated by "_"
//   - If a prefix is specified, it is added at the beginning: PREFIX_KEY
//   - Example: for structure Server.Host with prefix "APP_" -> APP_SERVER_HOST
//
// Example:
//
//	type Config struct {
//	    Server struct {
//	        Host string `koanf:"host"`
//	        Port int    `koanf:"port"`
//	    } `koanf:"server"`
//	}
//
//	var cfg Config
//	// Override via: APP_SERVER_HOST, APP_SERVER_PORT
//	config.LoadWithPrefixDefault(&cfg, "APP_")
func LoadWithPrefixDefault(target any, envPrefix string) {
	configPath := getDefaultConfigPath()
	if err := LoadWithPrefix(configPath, target, envPrefix); err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
}

// LoadSection loads a specific section from a YAML file with override via environment variables.
// Useful when configurations for multiple services are stored in one YAML file.
//
// Parameters:
//   - configPath: path to the YAML configuration file
//   - section: section name in the YAML file (e.g., "database", "redis")
//   - target: pointer to the structure into which the configuration will be loaded
//   - envPrefix: prefix for environment variables
//
// Example YAML:
//
//	database:
//	  host: localhost
//	  port: 5432
//	redis:
//	  host: localhost
//	  port: 6379
//
// Example usage:
//
//	type DBConfig struct {
//	    Host string `koanf:"host"`
//	    Port int    `koanf:"port"`
//	}
//
//	var dbCfg DBConfig
//	// Override via: DB_HOST, DB_PORT
//	err := config.LoadSection("config.yaml", "database", &dbCfg, "DB")
func LoadSection(configPath string, section string, target any, envPrefix string) error {
	k := koanf.New(".")

	// 1. Load configuration from YAML file
	if err := k.Load(file.Provider(configPath), yaml.Parser()); err != nil {
		return fmt.Errorf("error loading configuration from file %s: %w", configPath, err)
	}

	// 2. Override with values from environment variables (if prefix is specified)
	// Callback function to transform environment variable names into configuration keys
	if envPrefix != "" {
		envCb := func(s string) string {
			// Remove prefix
			if after, ok := strings.CutPrefix(s, envPrefix); ok {
				s = after
			}
			// Transform HOST -> host and add section: section.host
			key := strings.ToLower(s)
			key = strings.ReplaceAll(key, "_", ".")
			return section + "." + key
		}

		if err := k.Load(env.Provider("", ".", envCb), nil); err != nil {
			return fmt.Errorf("error loading environment variables: %w", err)
		}
	}

	// 3. Unmarshal specific section into target structure
	if err := k.Unmarshal(section, target); err != nil {
		return fmt.Errorf("error deserializing section '%s': %w", section, err)
	}

	return nil
}

// LoadSectionDefault loads a specific section from the default config.yaml file (next to the executable)
// with override via environment variables.
// Useful when configurations for multiple services are stored in one YAML file.
//
// Parameters:
//   - section: section name in the YAML file (e.g., "database", "redis")
//   - target: pointer to the structure into which the configuration will be loaded
//   - envPrefix: prefix for environment variables
//
// Example YAML:
//
//	database:
//	  host: localhost
//	  port: 5432
//	redis:
//	  host: localhost
//	  port: 6379
//
// Example usage:
//
//	type DBConfig struct {
//	    Host string `koanf:"host"`
//	    Port int    `koanf:"port"`
//	}
//
//	var dbCfg DBConfig
//	// Override via: DB_HOST, DB_PORT
//	err := config.LoadSectionDefault("database", &dbCfg, "DB_")
func LoadSectionDefault(section string, target any, envPrefix string) error {
	configPath := getDefaultConfigPath()
	return LoadSection(configPath, section, target, envPrefix)
}
