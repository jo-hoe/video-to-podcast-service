package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadUIConfig(t *testing.T) {
	// Test with default values
	cfg := LoadUIConfig()

	if cfg.Server.Port != "3000" {
		t.Errorf("Expected default port 3000, got %s", cfg.Server.Port)
	}

	if cfg.API.BaseURL != "http://api-service:8080" {
		t.Errorf("Expected default API BaseURL http://api-service:8080, got %s", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", cfg.API.Timeout)
	}

	// Test with environment variables
	_ = os.Setenv(UIEnvPort, "4000")
	_ = os.Setenv(UIEnvAPIBaseURL, "http://test-api:8080")
	_ = os.Setenv(UIEnvAPITimeout, "45s")
	defer func() {
		_ = os.Unsetenv(UIEnvPort)
		_ = os.Unsetenv(UIEnvAPIBaseURL)
		_ = os.Unsetenv(UIEnvAPITimeout)
	}()

	cfg = LoadUIConfig()

	if cfg.Server.Port != "4000" {
		t.Errorf("Expected port 4000 from env, got %s", cfg.Server.Port)
	}

	if cfg.API.BaseURL != "http://test-api:8080" {
		t.Errorf("Expected API BaseURL http://test-api:8080 from env, got %s", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s from env, got %v", cfg.API.Timeout)
	}
}

func TestLoadUIConfigInvalidTimeout(t *testing.T) {
	// Test with invalid timeout format
	_ = os.Setenv(UIEnvAPITimeout, "invalid")
	defer func() { _ = os.Unsetenv(UIEnvAPITimeout) }()

	cfg := LoadUIConfig()

	// Should fallback to default timeout when parsing fails
	if cfg.API.Timeout != 30*time.Second {
		t.Errorf("Expected fallback timeout 30s for invalid format, got %v", cfg.API.Timeout)
	}
}
