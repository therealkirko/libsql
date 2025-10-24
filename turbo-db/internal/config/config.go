package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	// Server configuration
	ServerPort     string
	ServerHost     string

	// Database configuration
	MetaDBPath     string
	DatabasesDir   string

	// LibSQL configuration
	LibSQLBinary   string
	PortRangeStart int
	PortRangeEnd   int

	// Authentication
	APIKey         string
	EnableAuth     bool
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() (*Config, error) {
	cfg := &Config{
		ServerPort:     getEnv("SERVER_PORT", "8080"),
		ServerHost:     getEnv("SERVER_HOST", "0.0.0.0"),
		MetaDBPath:     getEnv("META_DB_PATH", "./turbo-meta.db"),
		DatabasesDir:   getEnv("DATABASES_DIR", "./databases"),
		LibSQLBinary:   getEnv("LIBSQL_BINARY", "../target/release/sqld"),
		PortRangeStart: getEnvInt("PORT_RANGE_START", 9000),
		PortRangeEnd:   getEnvInt("PORT_RANGE_END", 9100),
		APIKey:         getEnv("API_KEY", ""),
		EnableAuth:     getEnvBool("ENABLE_AUTH", false),
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.PortRangeStart >= c.PortRangeEnd {
		return fmt.Errorf("PORT_RANGE_START must be less than PORT_RANGE_END")
	}

	if c.EnableAuth && c.APIKey == "" {
		return fmt.Errorf("API_KEY must be set when ENABLE_AUTH is true")
	}

	return nil
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt gets an environment variable as int or returns a default value
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvBool gets an environment variable as bool or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}
