package config

import (
	"time"
)

// UIConfig holds all UI service configuration
type UIConfig struct {
	Server UIServerConfig
	API    APIClientConfig
}

// UIServerConfig holds UI server-related configuration
type UIServerConfig struct {
	Port string
}

// APIClientConfig holds API client configuration for communicating with API service
type APIClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

// Environment variable names for UI service
const (
	UIEnvPort       = "UI_PORT"
	UIEnvAPIBaseURL = "API_BASE_URL"
	UIEnvAPITimeout = "API_TIMEOUT"
)

// LoadUIConfig loads UI service configuration from environment variables with sensible defaults
func LoadUIConfig() *UIConfig {
	timeoutStr := getEnvOrDefault(UIEnvAPITimeout, "30s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		// If parsing fails, use default timeout
		timeout = 30 * time.Second
	}

	return &UIConfig{
		Server: UIServerConfig{
			Port: getEnvOrDefault(UIEnvPort, "3000"),
		},
		API: APIClientConfig{
			BaseURL: getEnvOrDefault(UIEnvAPIBaseURL, "http://api-service:8080"),
			Timeout: timeout,
		},
	}
}
