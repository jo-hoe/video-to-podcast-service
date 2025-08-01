package config

import (
	"os"
	"path/filepath"
)

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

// getEnvOrDefault returns the value of an environment variable or a default value if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDefaultResourcePath returns the default path for storing resources
func getDefaultResourcePath() string {
	// Check if we're in a containerized environment with common data directory
	if _, err := os.Stat("/app/data"); err == nil {
		return "/app/data/resources"
	}

	// Use current working directory + resources as default for local development
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path if we can't get current directory
		return "resources"
	}
	return filepath.Join(cwd, "resources")
}

// getDatabaseConnectionString returns the default database connection string
func getDatabaseConnectionString() string {
	// Check for DATABASE_PATH environment variable first
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		return filepath.Join(dbPath, "podcast_items.db")
	}

	// Check if we're in a containerized environment with common data directory
	if _, err := os.Stat("/app/data"); err == nil {
		// Ensure database directory exists
		dbDir := "/app/data/database"
		if err := os.MkdirAll(dbDir, 0755); err == nil {
			return filepath.Join(dbDir, "podcast_items.db")
		}
	}

	// Fallback to current directory for local development
	return "podcast_items.db"
}
