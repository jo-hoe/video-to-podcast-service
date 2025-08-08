package api

import (
	"bytes"
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
	"github.com/jo-hoe/video-to-podcast-service/internal/server/ui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TestAPIClientCommunication tests the API client communication with various scenarios
func TestAPIClientCommunication(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "integration_test_*")
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

	// Create core service
	coreService := core.NewCoreService(databaseService, tempDir)

	// Create API service
	apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	// Create Echo instance for API
	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	// Create test server
	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	// Create API client
	apiClient := NewAPIClient(apiServer.URL, 5*time.Second)

	t.Run("HealthCheck", func(t *testing.T) {
		err := apiClient.HealthCheck()
		if err != nil {
			t.Errorf("HealthCheck failed: %v", err)
		}
	})

	t.Run("GetAllPodcastItems_Empty", func(t *testing.T) {
		items, err := apiClient.GetAllPodcastItems()
		if err != nil {
			t.Errorf("GetAllPodcastItems failed: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(items))
		}
	})

	t.Run("GetFeeds_Empty", func(t *testing.T) {
		feeds, err := apiClient.GetFeeds()
		if err != nil {
			t.Errorf("GetFeeds failed: %v", err)
		}
		if len(feeds) != 0 {
			t.Errorf("Expected 0 feeds, got %d", len(feeds))
		}
	})

	t.Run("AddItems_InvalidURL", func(t *testing.T) {
		err := apiClient.AddItems([]string{"invalid-url"})
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}
	})

	t.Run("DeletePodcastItem_NotFound", func(t *testing.T) {
		err := apiClient.DeletePodcastItem("nonexistent", "nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent item, got nil")
		}
	})
}

// TestEndToEndWorkflow tests the complete workflow between UI and API services
func TestEndToEndWorkflow(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "e2e_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up API service
	dbPath := filepath.Join(tempDir, "test.db")
	databaseService, err := database.NewDatabase(dbPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	coreService := core.NewCoreService(databaseService, tempDir)
	apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	// Set up UI service
	apiClient := NewAPIClient(apiServer.URL, 5*time.Second)
	uiService := ui.NewUIService(apiClient)

	uiEcho := echo.New()
	uiEcho.Use(middleware.Logger())
	uiEcho.Use(middleware.Recover())
	uiService.SetUIRoutes(uiEcho)

	uiServer := httptest.NewServer(uiEcho)
	defer uiServer.Close()

	t.Run("UI_Index_Page", func(t *testing.T) {
		resp, err := http.Get(uiServer.URL + "/")
		if err != nil {
			t.Fatalf("Failed to get index page: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		// Check that the page contains expected elements
		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Video to Podcast") {
			t.Error("Index page should contain 'Video to Podcast' title")
		}
	})

	t.Run("API_Health_Check", func(t *testing.T) {
		resp, err := http.Get(apiServer.URL + "/v1/health")
		if err != nil {
			t.Fatalf("Failed to get health check: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("API_Get_Empty_Feeds", func(t *testing.T) {
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

		if len(feeds) != 0 {
			t.Errorf("Expected 0 feeds, got %d", len(feeds))
		}
	})

	t.Run("API_Get_Empty_Items", func(t *testing.T) {
		resp, err := http.Get(apiServer.URL + "/v1/items")
		if err != nil {
			t.Fatalf("Failed to get items: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}

		var items []*database.PodcastItem
		if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
			t.Fatalf("Failed to decode items response: %v", err)
		}

		if len(items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(items))
		}
	})
}

// TestRSSFeedConfigurationIntegration tests comprehensive RSS feed configuration scenarios
func TestRSSFeedConfigurationIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "feed_config_integration_*")
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

	// Create test audio files and podcast items
	testItems := []*database.PodcastItem{
		{
			ID:                     "tech1",
			Title:                  "Tech Talk Episode 1",
			Description:            "Latest technology trends",
			AudioFilePath:          filepath.Join(tempDir, "tech-channel", "tech1.mp3"),
			Author:                 "Tech Expert",
			DurationInMilliseconds: 1800000, // 30 minutes
			Thumbnail:              "http://example.com/tech1.jpg",
		},
		{
			ID:                     "tech2",
			Title:                  "Tech Talk Episode 2",
			Description:            "AI and Machine Learning",
			AudioFilePath:          filepath.Join(tempDir, "tech-channel", "tech2.mp3"),
			Author:                 "Tech Expert",
			DurationInMilliseconds: 2100000, // 35 minutes
			Thumbnail:              "http://example.com/tech2.jpg",
		},
		{
			ID:                     "music1",
			Title:                  "Music Review: Album X",
			Description:            "Review of the latest album",
			AudioFilePath:          filepath.Join(tempDir, "music-channel", "music1.mp3"),
			Author:                 "Music Critic",
			DurationInMilliseconds: 1200000, // 20 minutes
			Thumbnail:              "http://example.com/music1.jpg",
		},
		{
			ID:                     "news1",
			Title:                  "Daily News Update",
			Description:            "Today's top stories",
			AudioFilePath:          filepath.Join(tempDir, "news-channel", "news1.mp3"),
			Author:                 "News Anchor",
			DurationInMilliseconds: 900000, // 15 minutes
			Thumbnail:              "http://example.com/news1.jpg",
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

	t.Run("PerDirectoryMode_CompleteFlow", func(t *testing.T) {
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

	t.Run("UnifiedMode_CompleteFlow", func(t *testing.T) {
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

	t.Run("DefaultModeBackwardCompatibility", func(t *testing.T) {
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

	t.Run("InvalidModeHandling", func(t *testing.T) {
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

		if len(feeds) != 2 {
			t.Errorf("Expected 2 feeds with invalid mode (default behavior), got %d", len(feeds))
		}
	})
}

// TestAPIServiceErrorHandling tests error handling scenarios in the API service
func TestAPIServiceErrorHandling(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "error_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up API service
	dbPath := filepath.Join(tempDir, "test.db")
	databaseService, err := database.NewDatabase(dbPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	coreService := core.NewCoreService(databaseService, tempDir)
	apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	t.Run("AddItems_EmptyBody", func(t *testing.T) {
		resp, err := http.Post(apiServer.URL+"/v1/addItems", "application/json", bytes.NewBuffer([]byte("{}")))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("AddItems_InvalidJSON", func(t *testing.T) {
		resp, err := http.Post(apiServer.URL+"/v1/addItems", "application/json", bytes.NewBuffer([]byte("invalid json")))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("Feed_NotFound", func(t *testing.T) {
		resp, err := http.Get(apiServer.URL + "/v1/feeds/nonexistent/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("AudioFile_NotFound", func(t *testing.T) {
		resp, err := http.Get(apiServer.URL + "/v1/feeds/nonexistent/nonexistent.mp3")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("DeleteItem_NotFound", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodDelete, apiServer.URL+"/v1/feeds/nonexistent/nonexistent", nil)
		if err != nil {
			t.Fatalf("Failed to create request: %v", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})
}

// TestUIServiceWithAPIFailures tests UI service behavior when API service fails
func TestUIServiceWithAPIFailures(t *testing.T) {
	// Create a failing API server
	failingAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("API service is down"))
	}))
	defer failingAPIServer.Close()

	// Create API client pointing to failing server
	apiClient := NewAPIClient(failingAPIServer.URL, 1*time.Second)

	// Create UI service
	uiService := ui.NewUIService(apiClient)

	uiEcho := echo.New()
	uiEcho.Use(middleware.Logger())
	uiEcho.Use(middleware.Recover())
	uiService.SetUIRoutes(uiEcho)

	uiServer := httptest.NewServer(uiEcho)
	defer uiServer.Close()

	t.Run("UI_Index_With_Failed_API", func(t *testing.T) {
		resp, err := http.Get(uiServer.URL + "/")
		if err != nil {
			t.Fatalf("Failed to get index page: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// UI should still serve the page even if API is down
		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200 even with failed API, got %d", resp.StatusCode)
		}
	})

	t.Run("API_Client_Circuit_Breaker", func(t *testing.T) {
		// Make enough calls to open circuit breaker
		for i := 0; i < 6; i++ {
			_, err := apiClient.GetAllPodcastItems()
			if err == nil {
				t.Errorf("Expected error on call %d", i+1)
			}
		}

		// Next call should fail quickly due to circuit breaker
		start := time.Now()
		items, err := apiClient.GetAllPodcastItems()
		duration := time.Since(start)

		// Should return empty slice due to graceful degradation
		if items == nil || len(items) != 0 {
			t.Errorf("Expected empty slice from graceful degradation, got %v", items)
		}

		// Should have graceful degradation error
		if err == nil || !strings.Contains(err.Error(), "graceful degradation") {
			t.Errorf("Expected graceful degradation error, got %v", err)
		}

		// Should fail quickly due to circuit breaker
		if duration > 500*time.Millisecond {
			t.Errorf("Expected quick failure with circuit breaker, took %v", duration)
		}
	})
}

// TestServiceCommunicationProtocol tests the communication protocol between services
func TestServiceCommunicationProtocol(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protocol_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up API service
	dbPath := filepath.Join(tempDir, "test.db")
	databaseService, err := database.NewDatabase(dbPath, tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	coreService := core.NewCoreService(databaseService, tempDir)
	apiService := NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	t.Run("Content_Type_Headers", func(t *testing.T) {
		// Test JSON content type for feeds endpoint
		resp, err := http.Get(apiServer.URL + "/v1/feeds")
		if err != nil {
			t.Fatalf("Failed to get feeds: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			t.Errorf("Expected JSON content type, got %s", contentType)
		}
	})

	t.Run("CORS_Headers", func(t *testing.T) {
		// Test that API service handles CORS properly
		req, err := http.NewRequest("OPTIONS", apiServer.URL+"/v1/feeds", nil)
		if err != nil {
			t.Fatalf("Failed to create OPTIONS request: %v", err)
		}
		req.Header.Set("Origin", "http://localhost:3000")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send OPTIONS request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// Should not return error for OPTIONS request
		if resp.StatusCode >= 400 {
			t.Errorf("OPTIONS request failed with status %d", resp.StatusCode)
		}
	})

	t.Run("Error_Response_Format", func(t *testing.T) {
		// Test that error responses have consistent format
		resp, err := http.Get(apiServer.URL + "/v1/feeds/nonexistent/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Fatalf("Failed to read response body: %v", err)
		}

		// Error response should contain meaningful message
		bodyStr := string(body)
		if len(bodyStr) == 0 {
			t.Error("Error response should contain error message")
		}
	})
}

// Helper function to create API client (imported from cmd/ui package)
func NewAPIClient(baseURL string, timeout time.Duration) APIClient {
	return &HTTPAPIClient{
		baseURL:    baseURL,
		maxRetries: 3,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second),
	}
}

// APIClient interface (copied from cmd/ui package for testing)
type APIClient interface {
	AddItems(urls []string) error
	GetAllPodcastItems() ([]*database.PodcastItem, error)
	GetFeeds() ([]string, error)
	DeletePodcastItem(feedTitle, podcastItemID string) error
	GetLinkToFeed(host, feedsPath, filePath string) string
	HealthCheck() error
}

// HTTPAPIClient implementation (simplified for testing)
type HTTPAPIClient struct {
	baseURL        string
	httpClient     *http.Client
	maxRetries     int
	circuitBreaker *CircuitBreaker
}

// CircuitBreaker implementation (simplified for testing)
type CircuitBreaker struct {
	state        int
	failureCount int
	maxFailures  int
	timeout      time.Duration
}

func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:       0, // closed
		maxFailures: maxFailures,
		timeout:     timeout,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	if cb.state == 1 { // open
		return fmt.Errorf("circuit breaker is open")
	}

	err := fn()
	if err != nil {
		cb.failureCount++
		if cb.failureCount >= cb.maxFailures {
			cb.state = 1 // open
		}
		return err
	}

	cb.failureCount = 0
	cb.state = 0 // closed
	return nil
}

// Implement APIClient methods for testing
func (c *HTTPAPIClient) AddItems(urls []string) error {
	return c.circuitBreaker.Call(func() error {
		requestBody := map[string][]string{"urls": urls}
		jsonData, _ := json.Marshal(requestBody)

		resp, err := c.httpClient.Post(c.baseURL+"/v1/addItems", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return nil
	})
}

func (c *HTTPAPIClient) GetAllPodcastItems() ([]*database.PodcastItem, error) {
	var items []*database.PodcastItem

	err := c.circuitBreaker.Call(func() error {
		resp, err := c.httpClient.Get(c.baseURL + "/v1/items")
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		return json.NewDecoder(resp.Body).Decode(&items)
	})

	if err != nil && strings.Contains(err.Error(), "circuit breaker") {
		return []*database.PodcastItem{}, &GracefulDegradationError{
			Operation: "GetAllPodcastItems",
			Cause:     err,
		}
	}

	return items, err
}

func (c *HTTPAPIClient) GetFeeds() ([]string, error) {
	var feeds []string

	err := c.circuitBreaker.Call(func() error {
		resp, err := c.httpClient.Get(c.baseURL + "/v1/feeds")
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}

		return json.NewDecoder(resp.Body).Decode(&feeds)
	})

	if err != nil && strings.Contains(err.Error(), "circuit breaker") {
		return []string{}, &GracefulDegradationError{
			Operation: "GetFeeds",
			Cause:     err,
		}
	}

	return feeds, err
}

func (c *HTTPAPIClient) DeletePodcastItem(feedTitle, podcastItemID string) error {
	return c.circuitBreaker.Call(func() error {
		url := fmt.Sprintf("%s/v1/feeds/%s/%s", c.baseURL, feedTitle, podcastItemID)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			return err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return nil
	})
}

func (c *HTTPAPIClient) HealthCheck() error {
	return c.circuitBreaker.Call(func() error {
		resp, err := c.httpClient.Get(c.baseURL + "/v1/health")
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("HTTP %d", resp.StatusCode)
		}
		return nil
	})
}

func (c *HTTPAPIClient) GetLinkToFeed(host, feedsPath, filePath string) string {
	feedTitle := filepath.Base(filepath.Dir(filePath))
	return fmt.Sprintf("%s/%s/%s/rss.xml", c.baseURL, feedsPath, feedTitle)
}

// GracefulDegradationError for testing
type GracefulDegradationError struct {
	Operation string
	Cause     error
}

func (e *GracefulDegradationError) Error() string {
	return fmt.Sprintf("graceful degradation for %s: %v", e.Operation, e.Cause)
}
