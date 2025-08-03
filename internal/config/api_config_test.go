package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadAPIConfig(t *testing.T) {
	// Test with default configuration (no config file)
	tempDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	nonExistentConfigPath := filepath.Join(tempDir, "nonexistent.yaml")
	cfg, err := LoadAPIConfig(nonExistentConfigPath)
	if err != nil {
		t.Fatalf("Expected no error for non-existent config, got %v", err)
	}

	if cfg.Server.Port != "8080" {
		t.Errorf("Expected default port 8080, got %s", cfg.Server.Port)
	}

	if cfg.Server.BaseURL != "" {
		t.Errorf("Expected empty BaseURL by default, got %s", cfg.Server.BaseURL)
	}

	if cfg.Database.ConnectionString != ":memory:" {
		t.Errorf("Expected default database :memory:, got %s", cfg.Database.ConnectionString)
	}

	if cfg.Storage.BasePath != "/tmp/video-to-podcast" {
		t.Errorf("Expected default storage path /tmp/video-to-podcast, got %s", cfg.Storage.BasePath)
	}
}

func TestLoadAPIConfigFromFile(t *testing.T) {
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
    port: "9090"
    base_url: "http://test.com"
  database:
    connection_string: "/tmp/test.db"
  storage:
    base_path: "/tmp/test-storage"
  external:
    ytdlp_cookies_file: "/tmp/cookies.txt"
ui:
  server:
    port: "3000"
  api:
    base_url: "http://localhost:8080"
    timeout: "30s"
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write test config: %v", err)
	}

	cfg, err := LoadAPIConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if cfg.Server.Port != "9090" {
		t.Errorf("Expected port 9090 from config, got %s", cfg.Server.Port)
	}

	if cfg.Server.BaseURL != "http://test.com" {
		t.Errorf("Expected BaseURL http://test.com from config, got %s", cfg.Server.BaseURL)
	}

	if cfg.Database.ConnectionString != "/tmp/test.db" {
		t.Errorf("Expected database /tmp/test.db from config, got %s", cfg.Database.ConnectionString)
	}

	if cfg.Storage.BasePath != "/tmp/test-storage" {
		t.Errorf("Expected storage path /tmp/test-storage from config, got %s", cfg.Storage.BasePath)
	}

	if cfg.External.YTDLPCookiesFile != "/tmp/cookies.txt" {
		t.Errorf("Expected cookies file /tmp/cookies.txt from config, got %s", cfg.External.YTDLPCookiesFile)
	}
}
