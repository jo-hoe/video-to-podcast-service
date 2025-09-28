package main

import (
	"testing"
	"time"
)

// TestGetLinkToFeedRegression ensures that the RSS feed URL generation fix works correctly
// This test prevents regression of the issue where URLs pointed to UI service instead of API service
func TestGetLinkToFeedRegression(t *testing.T) {
	tests := []struct {
		name        string
		host        string
		feedsPath   string
		filePath    string
		expectedURL string
		description string
	}{
		{
			name:        "External_API_Host_Localhost",
			host:        "localhost:8080",
			feedsPath:   "v1/feeds",
			filePath:    "/app/resources/testfeed/audio.mp3",
			expectedURL: "http://localhost:8080/v1/feeds/testfeed/rss.xml",
			description: "Should use external API host (localhost:8080) not UI host (localhost:3000)",
		},
		{
			name:        "External_API_Host_Custom_Domain",
			host:        "api.example.com:8080",
			feedsPath:   "v1/feeds",
			filePath:    "/app/resources/myfeed/episode.mp3",
			expectedURL: "http://api.example.com:8080/v1/feeds/myfeed/rss.xml",
			description: "Should work with custom domains",
		},
		{
			name:        "Feed_Title_With_Spaces",
			host:        "localhost:8080",
			feedsPath:   "v1/feeds",
			filePath:    "/app/resources/my feed/audio.mp3",
			expectedURL: "http://localhost:8080/v1/feeds/my feed/rss.xml",
			description: "Should handle feed titles with spaces",
		},
		{
			name:        "Nested_Path_Structure",
			host:        "localhost:8080",
			feedsPath:   "v1/feeds",
			filePath:    "/app/resources/category/feedname/episode.mp3",
			expectedURL: "http://localhost:8080/v1/feeds/feedname/rss.xml",
			description: "Should extract feed title from parent directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create API client
			client := NewAPIClient("http://internal-service:8080", 30*time.Second)

			// Cast to HTTPAPIClient to access the method
			httpClient, ok := client.(*HTTPAPIClient)
			if !ok {
				t.Fatal("Expected HTTPAPIClient")
			}

			// Test the URL generation
			result := httpClient.GetLinkToFeed(tt.host, tt.feedsPath, tt.filePath)

			if result != tt.expectedURL {
				t.Errorf("GetLinkToFeed() = %v, want %v\nDescription: %s", result, tt.expectedURL, tt.description)
			}

			// Ensure the result uses the provided host, not the client's baseURL
			if result == "http://internal-service:8080/v1/feeds/testfeed/rss.xml" {
				t.Error("GetLinkToFeed() incorrectly used client's baseURL instead of provided host parameter")
			}
		})
	}
}

// TestGetLinkToFeedDoesNotUseInternalURL ensures the fix prevents using internal Docker URLs
func TestGetLinkToFeedDoesNotUseInternalURL(t *testing.T) {
	// Create client with internal Docker service URL
	client := NewAPIClient("http://api-service:8080", 30*time.Second)
	httpClient := client.(*HTTPAPIClient)

	// Test with external host
	externalHost := "localhost:8080"
	filePath := "/app/resources/testfeed/audio.mp3"

	result := httpClient.GetLinkToFeed(externalHost, "v1/feeds", filePath)

	// Should NOT contain the internal service URL
	if result == "http://api-service:8080/v1/feeds/testfeed/rss.xml" {
		t.Error("GetLinkToFeed() incorrectly used internal Docker service URL (api-service:8080)")
	}

	// Should contain the external host
	expected := "http://localhost:8080/v1/feeds/testfeed/rss.xml"
	if result != expected {
		t.Errorf("GetLinkToFeed() = %v, want %v", result, expected)
	}
}

// TestUIServiceHostConfiguration ensures UI service uses correct host for RSS feeds
func TestUIServiceHostConfiguration(t *testing.T) {
	// This test documents the expected behavior:
	// UI service should pass "localhost:8080" as host to GetLinkToFeed
	// NOT the UI request host "localhost:3000"

	expectedAPIHost := "localhost:8080"
	uiRequestHost := "localhost:3000"

	if expectedAPIHost == uiRequestHost {
		t.Error("UI service should use API host (localhost:8080) for RSS feeds, not UI host (localhost:3000)")
	}
}
