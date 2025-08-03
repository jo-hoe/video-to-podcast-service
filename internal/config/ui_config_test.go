package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadUIConfig(t *testing.T) {
	// Test with default configuration (no config file)
	tempDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	nonExistentConfigPath := filepath.Join(tempDir, "nonexistent.yaml")
	cfg, err := LoadUIConfig(nonExistentConfigPath)
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got %v", err)
	}

	if cfg.Server.Port != "3000" {
		t.Errorf("Expected default port 3000, got %s", cfg.Server.Port)
	}

	if cfg.API.BaseURL != "http://localhost:8080" {
		t.Errorf("Expected default API BaseURL http://localhost:8080, got %s", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 30*time.Second {
		t.Errorf("Expected default timeout 30s, got %v", cfg.API.Timeout)
	}
}

func TestLoadUIConfigFromFile(t *testing.T) {
	// Test with custom YAML configuration
	tempDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "test_config.yaml")
	configContent := `
api:
  server:
    port: "8080"
    base_url: ""
  database:
    connection_string: ":memory:"
  storage:
    base_path: "/tmp/video-to-podcast"
  external:
    ytdlp_cookies_file: ""
ui:
  server:
    port: "4000"
  api:
    base_url: "http://test-api:8080"
    timeout: "45s"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadUIConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != "4000" {
		t.Errorf("Expected port 4000 from config, got %s", cfg.Server.Port)
	}

	if cfg.API.BaseURL != "http://test-api:8080" {
		t.Errorf("Expected API BaseURL http://test-api:8080 from config, got %s", cfg.API.BaseURL)
	}

	if cfg.API.Timeout != 45*time.Second {
		t.Errorf("Expected timeout 45s from config, got %v", cfg.API.Timeout)
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	// Test with invalid YAML content
	tempDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "invalid.yaml")
	// Create truly invalid YAML with syntax errors
	invalidContent := `
api:
  server:
    port: "8080"
    base_url: "
ui:
  server:
    port: [invalid structure without closing bracket
`

	if err := os.WriteFile(configPath, []byte(invalidContent), 0644); err != nil {
		t.Fatalf("Failed to write invalid config: %v", err)
	}

	_, err = LoadUIConfig(configPath)
	if err == nil {
		t.Error("Expected error for invalid YAML, got nil")
	}
}
