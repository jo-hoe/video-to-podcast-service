package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Port        int         `yaml:"port"`
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
	DataSource       string `yaml:"dataSource"`
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
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	// Parse YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configPath, err)
	}

	// Convert relative paths to absolute paths
	if err := makePathsAbsolute(&config, filepath.Dir(configPath)); err != nil {
		return nil, fmt.Errorf("failed to resolve paths: %w", err)
	}

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
func makePathsAbsolute(config *Config, basePath string) error {
	// Convert database data source path
	if !filepath.IsAbs(config.Persistence.Database.DataSource) {
		config.Persistence.Database.DataSource = filepath.Join(basePath, config.Persistence.Database.DataSource)
	}

	// Convert cookie path
	if !filepath.IsAbs(config.Persistence.Cookies.CookiePath) {
		config.Persistence.Cookies.CookiePath = filepath.Join(basePath, config.Persistence.Cookies.CookiePath)
	}

	// Convert media paths
	if !filepath.IsAbs(config.Persistence.Media.MediaPath) {
		config.Persistence.Media.MediaPath = filepath.Join(basePath, config.Persistence.Media.MediaPath)
	}

	if !filepath.IsAbs(config.Persistence.Media.TempPath) {
		config.Persistence.Media.TempPath = filepath.Join(basePath, config.Persistence.Media.TempPath)
	}

	// Ensure directories exist
	return createDirectories(config)
}

// createDirectories creates necessary directories if they don't exist
func createDirectories(config *Config) error {
	dirs := []string{
		filepath.Dir(config.Persistence.Database.DataSource),
		filepath.Dir(config.Persistence.Cookies.CookiePath),
		config.Persistence.Media.MediaPath,
		config.Persistence.Media.TempPath,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
