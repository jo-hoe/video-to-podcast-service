package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TestRSSFeedConfigurationComprehensive tests all RSS feed configuration scenarios
func TestRSSFeedConfigurationComprehensive(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "feed_config_comprehensive_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up test database with comprehensive test data
	dbPath := filepath.Join(tempDir, "test.db")
	databaseService, err := database.NewDatabase(dbPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create multiple test directories to simulate different channels
	testDirs := []string{"tech-channel", "music-channel", "news-channel"}
	for _, dir := range testDirs {
		fullPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test audio files and podcast items with different timestamps
	baseTime := time.Now().Add(-24 * time.Hour)
	testItems := []*database.PodcastItem{
		{
			ID:                     "tech1",
			Title:                  "Tech Talk Episode 1",
			Description:            "Latest technology trends",
			AudioFilePath:          filepath.Join(tempDir, "tech-channel", "tech1.mp3"),
			Author:                 "Tech Expert",
			DurationInMilliseconds: 1800000, // 30 minutes
			Thumbnail:              "http://example.com/tech1.jpg",
			CreatedAt:              baseTime.Add(1 * time.Hour),
		},
		{
			ID:                     "tech2",
			Title:                  "Tech Talk Episode 2",
			Description:            "AI and Machine Learning",
			AudioFilePath:          filepath.Join(tempDir, "tech-channel", "tech2.mp3"),
			Author:                 "Tech Expert",
			DurationInMilliseconds: 2100000, // 35 minutes
			Thumbnail:              "http://example.com/tech2.jpg",
			CreatedAt:              baseTime.Add(3 * time.Hour),
		},
		{
			ID:                     "music1",
			Title:                  "Music Review: Album X",
			Description:            "Review of the latest album",
			AudioFilePath:          filepath.Join(tempDir, "music-channel", "music1.mp3"),
			Author:                 "Music Critic",
			DurationInMilliseconds: 1200000, // 20 minutes
			Thumbnail:              "http://example.com/music1.jpg",
			CreatedAt:              baseTime.Add(2 * time.Hour),
		},
		{
			ID:                     "news1",
			Title:                  "Daily News Update",
			Description:            "Today's top stories",
			AudioFilePath:          filepath.Join(tempDir, "news-channel", "news1.mp3"),
			Author:                 "News Anchor",
			DurationInMilliseconds: 900000, // 15 minutes
			Thumbnail:              "http://example.com/news1.jpg",
			CreatedAt:              baseTime.Add(4 * time.Hour),
		},
	}

	// Create audio files and add items to database
	for _, item := range testItems {
		if err := os.WriteFile(item.AudioFilePath, []byte("fake audio data"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", item.AudioFilePath, err)
		}
		if err := databaseService.CreatePodcastItem(item); err != nil {
			t.Fatalf("Failed to add test item %s: %v", item.ID, err)
		}
	}

	// Test per-directory mode (default behavior)
	t.Run("PerDirectoryMode_DefaultBehavior", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test feeds endpoint returns multiple feeds (one per directory)
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		// Should have 3 feeds (tech-channel, music-channel, news-channel)
		if len(feeds) != 3 {
			t.Errorf("Expected 3 feeds in per_directory mode, got %d", len(feeds))
		}

		// Verify each feed URL contains the expected pattern
		expectedChannels := map[string]bool{"tech-channel": false, "music-channel": false, "news-channel": false}
		for _, feedURL := range feeds {
			for channel := range expectedChannels {
				if strings.Contains(feedURL, channel) {
					expectedChannels[channel] = true
				}
			}
		}

		for channel, found := range expectedChannels {
			if !found {
				t.Errorf("Expected to find feed for channel %s", channel)
			}
		}

		// Test individual feed access works for each channel
		for channel := range expectedChannels {
			resp, err = http.Get(apiServer.URL + "/v1/feeds/" + channel + "/rss.xml")
			if err != nil {
				t.Fatalf("Failed to get individual feed for %s: %v", channel, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for individual feed %s, got %d", channel, resp.StatusCode)
			}

			// Verify content type is XML
			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, "xml") {
				t.Errorf("Expected XML content type for RSS feed %s, got %s", channel, contentType)
			}

			// Verify RSS content contains expected elements
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read RSS body for %s: %v", channel, err)
			}
			rssContent := string(body)

			// Check for RSS structure
			if !strings.Contains(rssContent, "<rss") {
				t.Errorf("RSS feed for %s should contain <rss tag", channel)
			}
			if !strings.Contains(rssContent, "<channel>") {
				t.Errorf("RSS feed for %s should contain <channel> tag", channel)
			}
			if !strings.Contains(rssContent, "<item>") {
				t.Errorf("RSS feed for %s should contain <item> tag", channel)
			}
		}

		// Test that "all" endpoint returns 404 in per-directory mode
		resp, err = http.Get(apiServer.URL + "/v1/feeds/all/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for /all endpoint in per-directory mode, got %d", resp.StatusCode)
		}

		// Test audio file access works
		resp, err = http.Get(apiServer.URL + "/v1/feeds/tech-channel/tech1.mp3")
		if err != nil {
			t.Fatalf("Failed to get audio file: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for audio file access, got %d", resp.StatusCode)
		}
	})

	// Test unified mode with multiple subdirectories
	t.Run("UnifiedMode_MultipleSubdirectories", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "unified"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test feeds endpoint returns single unified feed
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 1 {
			t.Errorf("Expected 1 feed in unified mode, got %d", len(feeds))
		}

		// Verify the unified feed URL contains expected pattern
		if len(feeds) > 0 && !strings.Contains(feeds[0], "/v1/feeds/all/rss.xml") {
			t.Errorf("Expected unified feed URL to contain '/v1/feeds/all/rss.xml', got %s", feeds[0])
		}

		// Test unified feed access works via /all endpoint
		resp, err = http.Get(apiServer.URL + "/v1/feeds/all/rss.xml")
		if err != nil {
			t.Fatalf("Failed to get unified feed via /all: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for unified feed via /all, got %d", resp.StatusCode)
		}

		// Verify content type is XML
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "xml") {
			t.Errorf("Expected XML content type for RSS feed, got %s", contentType)
		}

		// Verify RSS content contains all items from all subdirectories
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read RSS body: %v", err)
		}
		rssContent := string(body)

		// Check for RSS structure
		if !strings.Contains(rssContent, "<rss") {
			t.Error("Unified RSS feed should contain <rss tag")
		}
		if !strings.Contains(rssContent, "<channel>") {
			t.Error("Unified RSS feed should contain <channel> tag")
		}
		if !strings.Contains(rssContent, "All Podcast Items") {
			t.Error("Unified RSS feed should contain 'All Podcast Items' title")
		}

		// Verify all test items are included in the unified feed
		expectedTitles := []string{"Tech Talk Episode 1", "Tech Talk Episode 2", "Music Review: Album X", "Daily News Update"}
		for _, title := range expectedTitles {
			if !strings.Contains(rssContent, title) {
				t.Errorf("Unified RSS feed should contain item: %s", title)
			}
		}

		// Test individual feed access returns 404 in unified mode for all channels
		testChannels := []string{"tech-channel", "music-channel", "news-channel"}
		for _, channel := range testChannels {
			resp, err = http.Get(apiServer.URL + "/v1/feeds/" + channel + "/rss.xml")
			if err != nil {
				t.Fatalf("Failed to send request for %s: %v", channel, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected status 404 for individual feed %s in unified mode, got %d", channel, resp.StatusCode)
			}
		}

		// Test audio file access still works in unified mode
		resp, err = http.Get(apiServer.URL + "/v1/feeds/tech-channel/tech1.mp3")
		if err != nil {
			t.Fatalf("Failed to get audio file: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for audio file access in unified mode, got %d", resp.StatusCode)
		}
	})

	// Test backward compatibility with existing configurations
	t.Run("BackwardCompatibility_NilConfig", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		// Test with nil config (should default to per_directory)
		apiService := NewAPIService(coreService, "8080", nil)

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Should behave like per_directory mode
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 3 {
			t.Errorf("Expected 3 feeds with nil config (default behavior), got %d", len(feeds))
		}

		// Test that individual feed access works (per-directory behavior)
		resp, err = http.Get(apiServer.URL + "/v1/feeds/tech-channel/rss.xml")
		if err != nil {
			t.Fatalf("Failed to get individual feed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for individual feed with nil config, got %d", resp.StatusCode)
		}

		// Test that "all" endpoint returns 404 (per-directory behavior)
		resp, err = http.Get(apiServer.URL + "/v1/feeds/all/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for /all endpoint with nil config, got %d", resp.StatusCode)
		}
	})

	// Test backward compatibility with empty mode
	t.Run("BackwardCompatibility_EmptyMode", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		// Test with empty mode (should default to per_directory)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: ""})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Should behave like per_directory mode
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 3 {
			t.Errorf("Expected 3 feeds with empty mode (default behavior), got %d", len(feeds))
		}
	})

	// Test error scenarios and HTTP status codes
	t.Run("ErrorScenarios_HTTPStatusCodes", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test 404 for non-existent feed in per-directory mode
		resp, err := http.Get(apiServer.URL + "/v1/feeds/nonexistent-channel/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for non-existent feed, got %d", resp.StatusCode)
		}

		// Test 404 for non-existent audio file
		resp, err = http.Get(apiServer.URL + "/v1/feeds/tech-channel/nonexistent.mp3")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for non-existent audio file, got %d", resp.StatusCode)
		}

		// Test 404 for "all" endpoint in per-directory mode
		resp, err = http.Get(apiServer.URL + "/v1/feeds/all/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for /all endpoint in per-directory mode, got %d", resp.StatusCode)
		}
	})

	// Test unified mode error scenarios
	t.Run("UnifiedMode_ErrorScenarios", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "unified"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test 404 for individual feeds in unified mode
		testChannels := []string{"tech-channel", "music-channel", "news-channel", "nonexistent-channel"}
		for _, channel := range testChannels {
			resp, err := http.Get(apiServer.URL + "/v1/feeds/" + channel + "/rss.xml")
			if err != nil {
				t.Fatalf("Failed to send request for %s: %v", channel, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected status 404 for individual feed %s in unified mode, got %d", channel, resp.StatusCode)
			}
		}

		// Test that unified feed access works
		resp, err := http.Get(apiServer.URL + "/v1/feeds/all/rss.xml")
		if err != nil {
			t.Fatalf("Failed to get unified feed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for unified feed access, got %d", resp.StatusCode)
		}
	})

	// Test invalid mode handling
	t.Run("InvalidMode_DefaultsToPerDirectory", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		// Test with invalid mode (should default to per_directory)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "invalid_mode"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Should behave like per_directory mode
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 3 {
			t.Errorf("Expected 3 feeds with invalid mode (default behavior), got %d", len(feeds))
		}

		// Test that individual feed access works (per-directory behavior)
		resp, err = http.Get(apiServer.URL + "/v1/feeds/tech-channel/rss.xml")
		if err != nil {
			t.Fatalf("Failed to get individual feed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for individual feed with invalid mode, got %d", resp.StatusCode)
		}

		// Test that "all" endpoint returns 404 (per-directory behavior)
		resp, err = http.Get(apiServer.URL + "/v1/feeds/all/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404 for /all endpoint with invalid mode, got %d", resp.StatusCode)
		}
	})
}

// TestConfigurationLoadingAndValidation tests configuration loading scenarios
func TestConfigurationLoadingAndValidation(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "config_validation_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up test database
	dbPath := filepath.Join(tempDir, "test.db")
	databaseService, err := database.NewDatabase(dbPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test directory and file
	testDir := filepath.Join(tempDir, "test-channel")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	testFile := filepath.Join(testDir, "test.mp3")
	if err := os.WriteFile(testFile, []byte("fake audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Add test podcast item
	testItem := &database.PodcastItem{
		ID:                     "test1",
		Title:                  "Test Item",
		Description:            "Test Description",
		AudioFilePath:          testFile,
		Author:                 "Test Author",
		DurationInMilliseconds: 60000,
		Thumbnail:              "http://example.com/test.jpg",
		CreatedAt:              time.Now(),
	}

	if err := databaseService.CreatePodcastItem(testItem); err != nil {
		t.Fatalf("Failed to add test item: %v", err)
	}

	t.Run("ValidPerDirectoryConfig", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test that per-directory mode works correctly
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 1 {
			t.Errorf("Expected 1 feed in per_directory mode, got %d", len(feeds))
		}

		// Test individual feed access works
		resp, err = http.Get(apiServer.URL + "/v1/feeds/test-channel/rss.xml")
		if err != nil {
			t.Fatalf("Failed to get individual feed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for individual feed, got %d", resp.StatusCode)
		}
	})

	t.Run("ValidUnifiedConfig", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "unified"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test that unified mode works correctly
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 1 {
			t.Errorf("Expected 1 feed in unified mode, got %d", len(feeds))
		}

		// Test unified feed access works
		resp, err = http.Get(apiServer.URL + "/v1/feeds/all/rss.xml")
		if err != nil {
			t.Fatalf("Failed to get unified feed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for unified feed, got %d", resp.StatusCode)
		}
	})

	t.Run("ConfigurationValidation_LoadAPIConfig", func(t *testing.T) {
		// Test configuration loading with valid modes
		validConfigs := []struct {
			mode     string
			expected string
		}{
			{"per_directory", "per_directory"},
			{"unified", "unified"},
			{"", "per_directory"}, // Empty should default to per_directory
		}

		for _, tc := range validConfigs {
			// Create temporary config file
			configPath := filepath.Join(tempDir, "test_config.yaml")
			configContent := fmt.Sprintf(`
api:
  feed:
    mode: "%s"
`, tc.mode)

			if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
				t.Fatalf("Failed to write config file: %v", err)
			}

			// Load configuration
			apiConfig, err := config.LoadAPIConfig(configPath)
			if err != nil {
				t.Fatalf("Failed to load API config: %v", err)
			}

			if apiConfig.Feed.Mode != tc.expected {
				t.Errorf("Expected feed mode %s, got %s", tc.expected, apiConfig.Feed.Mode)
			}

			// Clean up
			_ = os.Remove(configPath)
		}
	})

	t.Run("ConfigurationValidation_InvalidMode", func(t *testing.T) {
		// Test configuration loading with invalid mode
		configPath := filepath.Join(tempDir, "invalid_config.yaml")
		configContent := `
api:
  feed:
    mode: "invalid_mode"
`

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Load configuration
		apiConfig, err := config.LoadAPIConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load API config: %v", err)
		}

		// Should default to per_directory for invalid mode
		if apiConfig.Feed.Mode != "per_directory" {
			t.Errorf("Expected feed mode per_directory for invalid mode, got %s", apiConfig.Feed.Mode)
		}

		// Clean up
		_ = os.Remove(configPath)
	})

	t.Run("ConfigurationValidation_MissingFeedSection", func(t *testing.T) {
		// Test configuration loading without feed section
		configPath := filepath.Join(tempDir, "no_feed_config.yaml")
		configContent := `
api:
  server:
    port: "8080"
`

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			t.Fatalf("Failed to write config file: %v", err)
		}

		// Load configuration
		apiConfig, err := config.LoadAPIConfig(configPath)
		if err != nil {
			t.Fatalf("Failed to load API config: %v", err)
		}

		// Should default to per_directory when feed section is missing
		if apiConfig.Feed.Mode != "per_directory" {
			t.Errorf("Expected feed mode per_directory when feed section is missing, got %s", apiConfig.Feed.Mode)
		}

		// Clean up
		_ = os.Remove(configPath)
	})
}

// TestEndToEndWorkflowWithFeedConfiguration tests complete end-to-end workflow with different feed configurations
func TestEndToEndWorkflowWithFeedConfiguration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "e2e_feed_config_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up test database
	dbPath := filepath.Join(tempDir, "test.db")
	databaseService, err := database.NewDatabase(dbPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	// Create test directories and files
	testDirs := []string{"podcast1", "podcast2"}
	for _, dir := range testDirs {
		fullPath := filepath.Join(tempDir, dir)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatalf("Failed to create test directory %s: %v", dir, err)
		}
	}

	// Create test items
	testItems := []*database.PodcastItem{
		{
			ID:                     "p1e1",
			Title:                  "Podcast 1 Episode 1",
			Description:            "First episode of podcast 1",
			AudioFilePath:          filepath.Join(tempDir, "podcast1", "episode1.mp3"),
			Author:                 "Host 1",
			DurationInMilliseconds: 1800000,
			Thumbnail:              "http://example.com/p1e1.jpg",
			CreatedAt:              time.Now().Add(-2 * time.Hour),
		},
		{
			ID:                     "p2e1",
			Title:                  "Podcast 2 Episode 1",
			Description:            "First episode of podcast 2",
			AudioFilePath:          filepath.Join(tempDir, "podcast2", "episode1.mp3"),
			Author:                 "Host 2",
			DurationInMilliseconds: 2100000,
			Thumbnail:              "http://example.com/p2e1.jpg",
			CreatedAt:              time.Now().Add(-1 * time.Hour),
		},
	}

	// Create audio files and add items to database
	for _, item := range testItems {
		if err := os.WriteFile(item.AudioFilePath, []byte("fake audio data"), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", item.AudioFilePath, err)
		}
		if err := databaseService.CreatePodcastItem(item); err != nil {
			t.Fatalf("Failed to add test item %s: %v", item.ID, err)
		}
	}

	t.Run("EndToEnd_PerDirectoryMode", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test complete workflow: list feeds -> access individual feeds -> access audio files

		// 1. List all feeds
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 2 {
			t.Errorf("Expected 2 feeds, got %d", len(feeds))
		}

		// 2. Access each individual feed
		for _, feedURL := range feeds {
			resp, err := http.Get(feedURL)
			if err != nil {
				t.Fatalf("Failed to get feed %s: %v", feedURL, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200 for feed %s, got %d", feedURL, resp.StatusCode)
			}

			// Verify RSS content
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read RSS body: %v", err)
			}
			rssContent := string(body)

			if !strings.Contains(rssContent, "<rss") {
				t.Errorf("RSS feed should contain <rss tag")
			}
		}

		// 3. Access audio files
		resp, err = http.Get(apiServer.URL + "/v1/feeds/podcast1/episode1.mp3")
		if err != nil {
			t.Fatalf("Failed to get audio file: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for audio file, got %d", resp.StatusCode)
		}
	})

	t.Run("EndToEnd_UnifiedMode", func(t *testing.T) {
		coreService := core.NewCoreService(databaseService, tempDir)
		apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "unified"})

		apiEcho := echo.New()
		apiEcho.Use(middleware.Logger())
		apiEcho.Use(middleware.Recover())
		apiService.SetAPIRoutes(apiEcho)

		apiServer := httptest.NewServer(apiEcho)
		defer apiServer.Close()

		// Test complete workflow: list feeds -> access unified feed -> access audio files

		// 1. List all feeds (should be just one)
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		var feeds []string
		if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
			t.Fatalf("Failed to decode feeds response: %v", err)
		}

		if len(feeds) != 1 {
			t.Errorf("Expected 1 feed in unified mode, got %d", len(feeds))
		}

		// 2. Access the unified feed
		// The feed URL should be relative to the server, so construct the full URL
		unifiedFeedURL := apiServer.URL + "/v1/feeds/all/rss.xml"
		resp, err = http.Get(unifiedFeedURL)
		if err != nil {
			t.Fatalf("Failed to get unified feed: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for unified feed, got %d", resp.StatusCode)
		}

		// Verify RSS content contains both episodes
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read RSS body: %v", err)
		}
		rssContent := string(body)

		if !strings.Contains(rssContent, "Podcast 1 Episode 1") {
			t.Error("Unified RSS feed should contain Podcast 1 Episode 1")
		}
		if !strings.Contains(rssContent, "Podcast 2 Episode 1") {
			t.Error("Unified RSS feed should contain Podcast 2 Episode 1")
		}

		// 3. Access audio files (should still work)
		resp, err = http.Get(apiServer.URL + "/v1/feeds/podcast1/episode1.mp3")
		if err != nil {
			t.Fatalf("Failed to get audio file: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 for audio file in unified mode, got %d", resp.StatusCode)
		}
	})
}
