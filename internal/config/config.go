package config

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Port        int         `yaml:"port"`
	LogLevel    string      `yaml:"logLevel"`
	Persistence Persistence `yaml:"persistence"`
}

// Persistence holds all persistence-related configuration
type Persistence struct {
	Database Database `yaml:"database"`
	Cookies  Cookies  `yaml:"cookies"`
	Media    Media    `yaml:"media"`
}

// Database holds database configuration
type Database struct {
	Driver           string `yaml:"driver"`
	ConnectionString string `yaml:"connectionString"`
}

// Cookies holds cookie configuration
type Cookies struct {
	Enabled    bool   `yaml:"enabled"`
	CookiePath string `yaml:"cookiePath"`
}

// Media holds media storage configuration
type Media struct {
	MediaPath string `yaml:"mediaPath"`
	TempPath  string `yaml:"tempPath"`
}

var globalConfig *Config

// LoadConfig loads configuration from the specified YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Read the config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Set default values for missing configuration
	if err := setDefaults(&config); err != nil {
		return nil, fmt.Errorf("failed to set default values: %w", err)
	}

	// Convert relative paths to absolute paths
	if err := makePathsAbsolute(&config, filepath.Dir(configPath)); err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

	// Log the loaded configuration
	logLoadedConfig(&config)

	globalConfig = &config
	return &config, nil
}

// GetConfig returns the globally loaded configuration
func GetConfig() *Config {
	if globalConfig == nil {
		panic("configuration not loaded - call LoadConfig first")
	}
	return globalConfig
}

// makePathsAbsolute converts relative paths in the config to absolute paths
// Paths starting with "./" are resolved relative to the current working directory (project root)
func makePathsAbsolute(config *Config, basePath string) error {
	// Get current working directory (project root) for resolving ./mount paths
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	// Define path mappings for DRY conversion
	pathMappings := []*string{
		&config.Persistence.Cookies.CookiePath,
		&config.Persistence.Media.MediaPath,
		&config.Persistence.Media.TempPath,
	}

	// Convert all relative paths to absolute paths
	// Use project root (cwd) instead of config file location for ./mount paths
	for _, pathPtr := range pathMappings {
		*pathPtr = makeAbsolutePathFromRoot(*pathPtr, cwd)
	}

	// Convert connection string path if it's a file-based database
	config.Persistence.Database.ConnectionString = makeAbsoluteConnectionStringFromRoot(config.Persistence.Database.ConnectionString, cwd)

	// Ensure directories exist
	return createDirectories(config)
}

// makeAbsolutePathFromRoot converts a relative path to absolute path using project root
func makeAbsolutePathFromRoot(path, projectRoot string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(projectRoot, path)
}

// createDirectories creates necessary directories if they don't exist
func createDirectories(config *Config) error {
	// Define directory paths dynamically to avoid duplication
	// All paths are guaranteed to have values due to setDefaults()
	directoryPaths := []string{
		filepath.Dir(config.Persistence.Cookies.CookiePath),
		config.Persistence.Media.MediaPath,
		config.Persistence.Media.TempPath,
		filepath.Dir(extractPathFromConnectionString(config.Persistence.Database.ConnectionString)),
	}

	return createDirectoriesFromPaths(directoryPaths)
}

// makeAbsoluteConnectionStringFromRoot converts relative paths in connection strings to absolute paths using project root
func makeAbsoluteConnectionStringFromRoot(connectionString, projectRoot string) string {
	// Handle SQLite file: prefix (guaranteed to exist due to defaults)
	if strings.HasPrefix(connectionString, "file:") {
		filePath := connectionString[5:] // Remove "file:" prefix
		if !filepath.IsAbs(filePath) {
			return "file:" + filepath.Join(projectRoot, filePath)
		}
	}
	return connectionString
}

// extractPathFromConnectionString extracts the file path from a connection string
func extractPathFromConnectionString(connectionString string) string {
	// Handle SQLite file: prefix (guaranteed to exist due to defaults)
	if strings.HasPrefix(connectionString, "file:") {
		return connectionString[5:] // Remove "file:" prefix
	}
	return ""
}

// setDefaults sets default values for configuration fields that weren't specified
func setDefaults(config *Config) error {
	// Set default port if not specified
	if config.Port == 0 {
		config.Port = 8080
	}
	// Set default log level if not specified
	if strings.TrimSpace(config.LogLevel) == "" {
		config.LogLevel = "info"
	}

	// Set default database configuration if not specified
	if config.Persistence.Database.Driver == "" {
		config.Persistence.Database.Driver = "sqlite3"
	}
	if config.Persistence.Database.ConnectionString == "" {
		config.Persistence.Database.ConnectionString = filepath.Join("file:.", "mount", "database", "video-to-podcast-service.db")
	}

	// Set default cookie configuration if not specified
	if config.Persistence.Cookies.CookiePath == "" {
		config.Persistence.Cookies.CookiePath = filepath.Join(".", "mount", "cookies", "youtube-cookies.txt")
	}
	// Default cookies to enabled
	// Note: We can't distinguish between false and unset for bool, so we assume false means disabled

	// Set default media configuration if not specified
	if config.Persistence.Media.MediaPath == "" {
		config.Persistence.Media.MediaPath = filepath.Join(".", "mount", "resources", "media")
	}
	if config.Persistence.Media.TempPath == "" {
		config.Persistence.Media.TempPath = filepath.Join(".", "mount", "resources", "temp")
	}

	return nil
}

// logLoadedConfig logs the loaded configuration values
func logLoadedConfig(config *Config) {
	slog.Info("=== Configuration Loaded ===")
	slog.Info("Port", "value", config.Port)
	slog.Info("Log Level", "value", config.LogLevel)
	slog.Info("Database Driver", "value", config.Persistence.Database.Driver)
	slog.Info("Database Connection", "value", config.Persistence.Database.ConnectionString)
	slog.Info("Cookies Enabled", "value", config.Persistence.Cookies.Enabled)
	slog.Info("Cookie Path", "value", config.Persistence.Cookies.CookiePath)
	slog.Info("Media Path", "value", config.Persistence.Media.MediaPath)
	slog.Info("Temp Path", "value", config.Persistence.Media.TempPath)
	slog.Info("============================")
}

// createDirectoriesFromPaths creates directories from a list of paths
func createDirectoriesFromPaths(paths []string) error {
	for _, dir := range paths {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}
