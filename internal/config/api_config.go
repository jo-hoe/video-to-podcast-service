package config

import (
	"os"
)

// APIConfig holds all API service configuration
type APIConfig struct {
	Server   APIServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	External ExternalConfig
}

// APIServerConfig holds API server-related configuration
type APIServerConfig struct {
	Port    string
	BaseURL string
}

// Environment variable names for API service
const (
	APIEnvPort             = "PORT"
	APIEnvBaseURL          = "BASE_URL"
	APIEnvBasePath         = "BASE_PATH"
	APIEnvDatabasePath     = "DATABASE_PATH"
	APIEnvConnectionString = "CONNECTION_STRING"
	APIEnvYTDLPCookiesFile = "YTDLP_COOKIES_FILE"
)

// LoadAPIConfig loads API service configuration from environment variables with sensible defaults
func LoadAPIConfig() *APIConfig {
	return &APIConfig{
		Server: APIServerConfig{
			Port:    getEnvOrDefault(APIEnvPort, "8080"),
			BaseURL: os.Getenv(APIEnvBaseURL),
		},
		Database: DatabaseConfig{
			ConnectionString: getEnvOrDefault(APIEnvConnectionString, getDatabaseConnectionString()),
		},
		Storage: StorageConfig{
			BasePath: getEnvOrDefault(APIEnvBasePath, getDefaultResourcePath()),
		},
		External: ExternalConfig{
			YTDLPCookiesFile: os.Getenv(APIEnvYTDLPCookiesFile),
		},
	}
}
