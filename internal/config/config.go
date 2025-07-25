package config

import (
	"os"
	"path/filepath"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	External ExternalConfig
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port    string
	BaseURL string
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	ConnectionString string
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	BasePath string
}

// ExternalConfig holds external service configuration
type ExternalConfig struct {
	YTDLPCookiesFile string
}

// Environment variable names
const (
	EnvPort             = "PORT"
	EnvBaseURL          = "BASE_URL"
	EnvBasePath         = "BASE_PATH"
	EnvConnectionString = "CONNECTION_STRING"
	EnvYTDLPCookiesFile = "YTDLP_COOKIES_FILE"
)

// LoadConfig loads configuration from environment variables with sensible defaults
func LoadConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:    getEnvOrDefault(EnvPort, "8080"),
			BaseURL: os.Getenv(EnvBaseURL),
		},
		Database: DatabaseConfig{
			ConnectionString: os.Getenv(EnvConnectionString),
		},
		Storage: StorageConfig{
			BasePath: getEnvOrDefault(EnvBasePath, getDefaultResourcePath()),
		},
		External: ExternalConfig{
			YTDLPCookiesFile: os.Getenv(EnvYTDLPCookiesFile),
		},
	}
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDefaultResourcePath returns the default resource path relative to executable
func getDefaultResourcePath() string {
	ex, err := os.Executable()
	if err != nil {
		return "resources"
	}
	exPath := filepath.Dir(ex)
	return filepath.Join(exPath, "resources")
}
