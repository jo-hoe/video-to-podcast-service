package config

import (
	"os"
	"testing"
)

func TestLoadAPIConfig(t *testing.T) {
	// Test with default values
	cfg := LoadAPIConfig()

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
	}

	if cfg.Server.BaseURL != "" {
		t.Errorf("Expected empty BaseURL by default, got %s", cfg.Server.BaseURL)
	}

	// Test with environment variables
	_ = os.Setenv(APIEnvPort, "9090")
	_ = os.Setenv(APIEnvBaseURL, "http://test.com")
	defer func() {
		_ = os.Unsetenv(APIEnvPort)
		_ = os.Unsetenv(APIEnvBaseURL)
	}()

	cfg = LoadAPIConfig()

	if cfg.Server.Port != "9090" {
		t.Errorf("Expected port 9090 from env, got %s", cfg.Server.Port)
	}

	if cfg.Server.BaseURL != "http://test.com" {
		t.Errorf("Expected BaseURL http://test.com from env, got %s", cfg.Server.BaseURL)
	}
}
