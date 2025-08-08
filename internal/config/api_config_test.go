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

	if cfg.Feed.Mode != "per_directory" {
		t.Errorf("Expected default feed mode per_directory, got %s", cfg.Feed.Mode)
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
  feed:
    mode: "unified"
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

	if cfg.Feed.Mode != "unified" {
		t.Errorf("Expected feed mode unified from config, got %s", cfg.Feed.Mode)
	}
}
func TestLoadAPIConfigWithMissingFeedConfig(t *testing.T) {
	// Test backward compatibility when feed config is missing from YAML
	tempDir, err := os.MkdirTemp("", "config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	configPath := filepath.Join(tempDir, "test_config_no_feed.yaml")
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

	// Should default to per_directory mode when feed config is missing
	if cfg.Feed.Mode != "per_directory" {
		t.Errorf("Expected default feed mode per_directory when config missing, got %s", cfg.Feed.Mode)
	}
}

func TestLoadAPIConfigWithDifferentFeedModes(t *testing.T) {
	testCases := []struct {
		name         string
		feedMode     string
		expectedMode string
	}{
		{
			name:         "per_directory mode",
			feedMode:     "per_directory",
			expectedMode: "per_directory",
		},
		{
			name:         "unified mode",
			feedMode:     "unified",
			expectedMode: "unified",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
  feed:
    mode: "` + tc.feedMode + `"
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

			if cfg.Feed.Mode != tc.expectedMode {
				t.Errorf("Expected feed mode %s, got %s", tc.expectedMode, cfg.Feed.Mode)
			}
		})
	}
}

func TestLoadAPIConfigWithInvalidFeedMode(t *testing.T) {
	testCases := []struct {
		name         string
		feedMode     string
		expectedMode string
	}{
		{
			name:         "invalid mode defaults to per_directory",
			feedMode:     "invalid_mode",
			expectedMode: "per_directory",
		},
		{
			name:         "empty mode defaults to per_directory",
			feedMode:     "",
			expectedMode: "per_directory",
		},
		{
			name:         "random string defaults to per_directory",
			feedMode:     "random_string",
			expectedMode: "per_directory",
		},
		{
			name:         "case sensitive - wrong case defaults to per_directory",
			feedMode:     "PER_DIRECTORY",
			expectedMode: "per_directory",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
  feed:
    mode: "` + tc.feedMode + `"
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

			if cfg.Feed.Mode != tc.expectedMode {
				t.Errorf("Expected feed mode %s for invalid input %q, got %s", tc.expectedMode, tc.feedMode, cfg.Feed.Mode)
			}
		})
	}
}
