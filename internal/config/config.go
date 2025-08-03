package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all application configuration
type Config struct {
	API APIConfig `yaml:"api"`
	UI  UIConfig  `yaml:"ui"`
}

// APIConfig holds all API service configuration
type APIConfig struct {
	Server   APIServerConfig `yaml:"server"`
	Database DatabaseConfig  `yaml:"database"`
	Storage  StorageConfig   `yaml:"storage"`
	External ExternalConfig  `yaml:"external"`
}

// UIConfig holds all UI service configuration
type UIConfig struct {
	Server UIServerConfig  `yaml:"server"`
	API    APIClientConfig `yaml:"api"`
}

// APIServerConfig holds API server-related configuration
type APIServerConfig struct {
	Port    string `yaml:"port"`
	BaseURL string `yaml:"base_url"`
}

// UIServerConfig holds UI server-related configuration
type UIServerConfig struct {
	Port string `yaml:"port"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	ConnectionString string `yaml:"connection_string"`
}

// StorageConfig holds storage-related configuration
type StorageConfig struct {
	BasePath string `yaml:"base_path"`
}

// ExternalConfig holds external service configuration
type ExternalConfig struct {
	YTDLPCookiesFile string `yaml:"ytdlp_cookies_file"`
}

// APIClientConfig holds API client configuration for communicating with API service
type APIClientConfig struct {
	BaseURL string        `yaml:"base_url"`
	Timeout time.Duration `yaml:"timeout"`
}

// LoadConfig loads configuration from YAML file with fallback to defaults
func LoadConfig(configPath string) (*Config, error) {
	// Start with default configuration
	config := getDefaultConfig()

	// If config file exists, load and merge it
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}

		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse config file: %w", err)
		}
	}

	return config, nil
}

// LoadAPIConfig loads only API service configuration
func LoadAPIConfig(configPath string) (*APIConfig, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &config.API, nil
}

// LoadUIConfig loads only UI service configuration
func LoadUIConfig(configPath string) (*UIConfig, error) {
	config, err := LoadConfig(configPath)
	if err != nil {
		return nil, err
	}
	return &config.UI, nil
}

// getDefaultConfig returns the default configuration for out-of-the-box usage
func getDefaultConfig() *Config {
	return &Config{
		API: APIConfig{
			Server: APIServerConfig{
				Port:    "8080",
				BaseURL: "", // Empty means auto-detect
			},
			Database: DatabaseConfig{
				ConnectionString: ":memory:", // In-memory database for no persistence
			},
			Storage: StorageConfig{
				BasePath: "/tmp/video-to-podcast", // Temporary storage
			},
			External: ExternalConfig{
				YTDLPCookiesFile: "", // No cookies file by default
			},
		},
		UI: UIConfig{
			Server: UIServerConfig{
				Port: "3000",
			},
			API: APIClientConfig{
				BaseURL: "http://localhost:8080", // Default for local development
				Timeout: 30 * time.Second,
			},
		},
	}
}

// SaveDefaultConfig saves the default configuration to a file
func SaveDefaultConfig(configPath string) error {
	config := getDefaultConfig()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetConfigPath returns the default configuration file path
func GetConfigPath() string {
	// Check if we're in a containerized environment
	if _, err := os.Stat("/app"); err == nil {
		return "/app/config.yaml"
	}

	// For local development
	return "config.yaml"
}
