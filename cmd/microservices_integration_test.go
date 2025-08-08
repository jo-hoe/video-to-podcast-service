package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/api"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/ui"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// TestMicroservicesAPIClientCommunication tests comprehensive API client communication scenarios
func TestMicroservicesAPIClientCommunication(t *testing.T) {
	// Create temporary directory for test database
	tempDir, err := os.MkdirTemp("", "microservices_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up test database and API service
	databaseService, err := database.NewDatabase("", tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	coreService := core.NewCoreService(databaseService, tempDir)
	apiService := api.NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	// Create API server
	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	// Create API client
	apiClient := NewTestAPIClient(apiServer.URL, 5*time.Second)

	t.Run("HealthCheck_Success", func(t *testing.T) {
		err := apiClient.HealthCheck()
		if err != nil {
			t.Errorf("HealthCheck failed: %v", err)
		}
	})

	t.Run("GetAllPodcastItems_EmptyDatabase", func(t *testing.T) {
		items, err := apiClient.GetAllPodcastItems()
		if err != nil {
			t.Errorf("GetAllPodcastItems failed: %v", err)
		}
		if len(items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(items))
		}
	})

	t.Run("GetFeeds_EmptyDatabase", func(t *testing.T) {
		feeds, err := apiClient.GetFeeds()
		if err != nil {
			t.Errorf("GetFeeds failed: %v", err)
		}
		if len(feeds) != 0 {
			t.Errorf("Expected 0 feeds, got %d", len(feeds))
		}
	})

	t.Run("AddItems_InvalidURL_ErrorHandling", func(t *testing.T) {
		err := apiClient.AddItems([]string{"invalid-url"})
		if err == nil {
			t.Error("Expected error for invalid URL, got nil")
		}

		// Verify error contains meaningful information
		if !strings.Contains(err.Error(), "HTTP") {
			t.Errorf("Expected HTTP error information, got: %v", err)
		}
	})

	t.Run("DeletePodcastItem_NotFound", func(t *testing.T) {
		err := apiClient.DeletePodcastItem("nonexistent", "nonexistent")
		if err == nil {
			t.Error("Expected error for nonexistent item, got nil")
		}
	})

	t.Run("GetLinkToFeed_URLGeneration", func(t *testing.T) {
		link := apiClient.GetLinkToFeed("localhost:8080", "v1/feeds", "/path/to/testfeed/audio.mp3")
		expectedPattern := "/v1/feeds/testfeed/rss.xml"
		if !strings.Contains(link, expectedPattern) {
			t.Errorf("Expected link to contain %s, got %s", expectedPattern, link)
		}
	})
}

// TestMicroservicesEndToEndWorkflow tests complete workflows between UI and API services
func TestMicroservicesEndToEndWorkflow(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "e2e_microservices_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up API service
	databaseService, err := database.NewDatabase("", tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	coreService := core.NewCoreService(databaseService, tempDir)
	apiService := api.NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	// Set up UI service
	apiClient := NewTestAPIClient(apiServer.URL, 5*time.Second)
	uiService := ui.NewUIService(apiClient)

	uiEcho := echo.New()
	uiEcho.Use(middleware.Logger())
	uiEcho.Use(middleware.Recover())
	uiService.SetUIRoutes(uiEcho)

	uiServer := httptest.NewServer(uiEcho)
	defer uiServer.Close()

	t.Run("UI_IndexPage_LoadsSuccessfully", func(t *testing.T) {
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

		bodyStr := string(body)
		if !strings.Contains(bodyStr, "Video to Podcast") {
			t.Error("Index page should contain 'Video to Podcast' title")
		}
	})

	t.Run("API_HealthCheck_Accessible", func(t *testing.T) {
		resp, err := http.Get(apiServer.URL + "/v1/health")
		if err != nil {
			t.Fatalf("Failed to get health check: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("Expected status 200, got %d", resp.StatusCode)
		}
	})

	t.Run("API_GetFeeds_ReturnsEmptyList", func(t *testing.T) {
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

	t.Run("API_GetItems_ReturnsEmptyList", func(t *testing.T) {
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

	t.Run("ServiceCommunication_UIToAPI", func(t *testing.T) {
		// Test that UI service can successfully communicate with API service
		// by making a request through the UI service that requires API communication

		// This simulates the HTMX add item functionality
		requestBody := `{"urls": ["https://example.com/invalid-video"]}`
		resp, err := http.Post(
			uiServer.URL+"/htmx/addItem",
			"application/json",
			bytes.NewBuffer([]byte(requestBody)),
		)
		if err != nil {
			t.Fatalf("Failed to send HTMX request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		// We expect this to fail (invalid URL), but the communication should work
		// The important thing is that we get a proper HTTP response, not a connection error
		if resp.StatusCode == 0 {
			t.Error("Expected HTTP response from UI service, got connection error")
		}
	})
}

// TestMicroservicesErrorHandling tests comprehensive error handling scenarios
func TestMicroservicesErrorHandling(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "error_handling_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up API service
	databaseService, err := database.NewDatabase("", tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	coreService := core.NewCoreService(databaseService, tempDir)
	apiService := api.NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	t.Run("API_AddItems_EmptyBody", func(t *testing.T) {
		resp, err := http.Post(apiServer.URL+"/v1/addItems", "application/json", bytes.NewBuffer([]byte("{}")))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("API_AddItems_InvalidJSON", func(t *testing.T) {
		resp, err := http.Post(apiServer.URL+"/v1/addItems", "application/json", bytes.NewBuffer([]byte("invalid json")))
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", resp.StatusCode)
		}
	})

	t.Run("API_Feed_NotFound", func(t *testing.T) {
		resp, err := http.Get(apiServer.URL + "/v1/feeds/nonexistent/rss.xml")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("API_AudioFile_NotFound", func(t *testing.T) {
		resp, err := http.Get(apiServer.URL + "/v1/feeds/nonexistent/nonexistent.mp3")
		if err != nil {
			t.Fatalf("Failed to send request: %v", err)
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", resp.StatusCode)
		}
	})

	t.Run("API_DeleteItem_NotFound", func(t *testing.T) {
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

// TestMicroservicesUIWithAPIFailures tests UI service behavior when API service fails
func TestMicroservicesUIWithAPIFailures(t *testing.T) {
	// Create a failing API server
	failingAPIServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("API service is down"))
	}))
	defer failingAPIServer.Close()

	// Create API client pointing to failing server
	apiClient := NewTestAPIClient(failingAPIServer.URL, 1*time.Second)

	// Create UI service
	uiService := ui.NewUIService(apiClient)

	uiEcho := echo.New()
	uiEcho.Use(middleware.Logger())
	uiEcho.Use(middleware.Recover())
	uiService.SetUIRoutes(uiEcho)

	uiServer := httptest.NewServer(uiEcho)
	defer uiServer.Close()

	t.Run("UI_IndexPage_WithFailedAPI", func(t *testing.T) {
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

	t.Run("APIClient_CircuitBreaker_Behavior", func(t *testing.T) {
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

	t.Run("APIClient_GracefulDegradation_GetFeeds", func(t *testing.T) {
		// Open circuit breaker first
		for i := 0; i < 6; i++ {
			_, _ = apiClient.GetFeeds()
		}

		// Test graceful degradation for GetFeeds
		feeds, err := apiClient.GetFeeds()

		// Should return empty slice with graceful degradation error
		if feeds == nil || len(feeds) != 0 {
			t.Errorf("Expected empty slice from graceful degradation, got %v", feeds)
		}

		if err == nil || !strings.Contains(err.Error(), "graceful degradation") {
			t.Errorf("Expected graceful degradation error, got %v", err)
		}
	})
}

// TestMicroservicesCommunicationProtocol tests the communication protocol between services
func TestMicroservicesCommunicationProtocol(t *testing.T) {
	// Create temporary directory for test
	tempDir, err := os.MkdirTemp("", "protocol_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer func() { _ = os.RemoveAll(tempDir) }()

	// Set up API service
	databaseService, err := database.NewDatabase("", tempDir)
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	coreService := core.NewCoreService(databaseService, tempDir)
	apiService := api.NewAPIService(coreService, "8080", &config.FeedConfig{Mode: "per_directory"})

	apiEcho := echo.New()
	apiEcho.Use(middleware.Logger())
	apiEcho.Use(middleware.Recover())
	apiService.SetAPIRoutes(apiEcho)

	apiServer := httptest.NewServer(apiEcho)
	defer apiServer.Close()

	t.Run("ContentType_Headers", func(t *testing.T) {
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

	t.Run("ErrorResponse_Format", func(t *testing.T) {
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

	t.Run("RequestResponse_Cycle", func(t *testing.T) {
		// Test complete request-response cycle for different endpoints
		endpoints := []struct {
			method   string
			path     string
			expected int
		}{
			{"GET", "/v1/health", http.StatusOK},
			{"GET", "/v1/feeds", http.StatusOK},
			{"GET", "/v1/items", http.StatusOK},
			{"GET", "/v1/feeds/nonexistent/rss.xml", http.StatusNotFound},
		}

		for _, endpoint := range endpoints {
			t.Run(fmt.Sprintf("%s_%s", endpoint.method, endpoint.path), func(t *testing.T) {
				req, err := http.NewRequest(endpoint.method, apiServer.URL+endpoint.path, nil)
				if err != nil {
					t.Fatalf("Failed to create request: %v", err)
				}

				client := &http.Client{}
				resp, err := client.Do(req)
				if err != nil {
					t.Fatalf("Failed to send request: %v", err)
				}
				defer func() { _ = resp.Body.Close() }()

				if resp.StatusCode != endpoint.expected {
					t.Errorf("Expected status %d, got %d", endpoint.expected, resp.StatusCode)
				}
			})
		}
	})
}

// TestMicroservicesCircuitBreakerRecovery tests circuit breaker recovery scenarios
func TestMicroservicesCircuitBreakerRecovery(t *testing.T) {
	callCount := 0

	// Create API server that fails first few calls, then succeeds
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 10 { // Fail enough times to open circuit breaker
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Temporary failure"))
			return
		}
		// Start succeeding
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer apiServer.Close()

	// Create API client with shorter circuit breaker timeout for testing
	client := NewTestAPIClient(apiServer.URL, 1*time.Second).(*TestHTTPAPIClient)
	client.circuitBreaker = NewTestCircuitBreaker(5, 100*time.Millisecond) // Faster recovery

	// Make calls to open circuit breaker
	for i := 0; i < 6; i++ {
		_, _ = client.GetAllPodcastItems()
	}

	// Verify circuit is open
	if client.circuitBreaker.GetState() != TestCircuitOpen {
		t.Errorf("Expected circuit to be open")
	}

	// Wait for circuit breaker timeout
	time.Sleep(150 * time.Millisecond)

	// Reset call count to ensure server starts succeeding
	callCount = 15

	// Make successful calls to close circuit
	for i := 0; i < 3; i++ {
		_, err := client.GetAllPodcastItems()
		if err != nil {
			// Check if it's a graceful degradation error (expected during recovery)
			if !strings.Contains(err.Error(), "graceful degradation") {
				t.Errorf("Expected success or graceful degradation after recovery, got %v", err)
			}
		}
	}

	// After successful calls, circuit should eventually close
	// We'll check the state after a few more successful calls
	for i := 0; i < 5; i++ {
		_, _ = client.GetAllPodcastItems()
		if client.circuitBreaker.GetState() == TestCircuitClosed {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	// The circuit should be closed or at least not permanently open
	if client.circuitBreaker.GetState() == TestCircuitOpen {
		t.Errorf("Circuit breaker should have recovered from open state")
	}
}

// Helper functions and test implementations

// TestAPIClient interface for testing (matches the real APIClient interface)
type TestAPIClient interface {
	AddItems(urls []string) error
	GetAllPodcastItems() ([]*database.PodcastItem, error)
	GetFeeds() ([]string, error)
	DeletePodcastItem(feedTitle, podcastItemID string) error
	GetLinkToFeed(host, feedsPath, filePath string) string
	HealthCheck() error
}

// TestCircuitBreakerState represents the state of the test circuit breaker
type TestCircuitBreakerState int

const (
	TestCircuitClosed TestCircuitBreakerState = iota
	TestCircuitOpen
	TestCircuitHalfOpen
)

// TestCircuitBreaker implements the circuit breaker pattern for testing
type TestCircuitBreaker struct {
	state            TestCircuitBreakerState
	failureCount     int
	lastFailureTime  time.Time
	successCount     int
	maxFailures      int
	timeout          time.Duration
	halfOpenMaxCalls int
}

// NewTestCircuitBreaker creates a new test circuit breaker
func NewTestCircuitBreaker(maxFailures int, timeout time.Duration) *TestCircuitBreaker {
	return &TestCircuitBreaker{
		state:            TestCircuitClosed,
		maxFailures:      maxFailures,
		timeout:          timeout,
		halfOpenMaxCalls: 3,
	}
}

// Call executes the given function with circuit breaker protection
func (cb *TestCircuitBreaker) Call(fn func() error) error {
	// Check if circuit should transition from open to half-open
	if cb.state == TestCircuitOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = TestCircuitHalfOpen
		cb.successCount = 0
	}

	// Reject calls if circuit is open
	if cb.state == TestCircuitOpen {
		return fmt.Errorf("circuit breaker is open")
	}

	// Limit calls in half-open state
	if cb.state == TestCircuitHalfOpen && cb.successCount >= cb.halfOpenMaxCalls {
		return fmt.Errorf("circuit breaker is half-open, max calls exceeded")
	}

	// Execute the function
	err := fn()

	if err != nil {
		cb.onFailure()
		return err
	}

	cb.onSuccess()
	return nil
}

func (cb *TestCircuitBreaker) onSuccess() {
	cb.failureCount = 0

	if cb.state == TestCircuitHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenMaxCalls {
			cb.state = TestCircuitClosed
		}
	}
}

func (cb *TestCircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == TestCircuitClosed && cb.failureCount >= cb.maxFailures {
		cb.state = TestCircuitOpen
	} else if cb.state == TestCircuitHalfOpen {
		cb.state = TestCircuitOpen
	}
}

// GetState returns the current state of the circuit breaker
func (cb *TestCircuitBreaker) GetState() TestCircuitBreakerState {
	return cb.state
}

// TestHTTPAPIClient implements TestAPIClient for testing
type TestHTTPAPIClient struct {
	baseURL        string
	httpClient     *http.Client
	maxRetries     int
	circuitBreaker *TestCircuitBreaker
}

// NewTestAPIClient creates a new test API client
func NewTestAPIClient(baseURL string, timeout time.Duration) TestAPIClient {
	return &TestHTTPAPIClient{
		baseURL:    baseURL,
		maxRetries: 3,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		circuitBreaker: NewTestCircuitBreaker(5, 30*time.Second),
	}
}

// Implement TestAPIClient methods

func (c *TestHTTPAPIClient) AddItems(urls []string) error {
	requestBody := map[string][]string{"urls": urls}
	jsonData, _ := json.Marshal(requestBody)

	return c.circuitBreaker.Call(func() error {
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

func (c *TestHTTPAPIClient) GetAllPodcastItems() ([]*database.PodcastItem, error) {
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
		return []*database.PodcastItem{}, &TestGracefulDegradationError{
			Operation: "GetAllPodcastItems",
			Cause:     err,
		}
	}

	return items, err
}

func (c *TestHTTPAPIClient) GetFeeds() ([]string, error) {
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
		return []string{}, &TestGracefulDegradationError{
			Operation: "GetFeeds",
			Cause:     err,
		}
	}

	return feeds, err
}

func (c *TestHTTPAPIClient) DeletePodcastItem(feedTitle, podcastItemID string) error {
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

func (c *TestHTTPAPIClient) HealthCheck() error {
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

func (c *TestHTTPAPIClient) GetLinkToFeed(host, feedsPath, filePath string) string {
	feedTitle := filepath.Base(filepath.Dir(filePath))
	return fmt.Sprintf("%s/%s/%s/rss.xml", c.baseURL, feedsPath, feedTitle)
}

// TestGracefulDegradationError for testing
type TestGracefulDegradationError struct {
	Operation string
	Cause     error
}

func (e *TestGracefulDegradationError) Error() string {
	return fmt.Sprintf("graceful degradation for %s: %v", e.Operation, e.Cause)
}

// TestServiceStartupRegression ensures that both services start correctly and key functionality works
// This test prevents regression of service startup issues and URL generation problems
func TestServiceStartupRegression(t *testing.T) {
	// Skip if not in CI or if explicitly requested
	if os.Getenv("SKIP_REGRESSION_TESTS") == "true" {
		t.Skip("Regression tests skipped")
	}

	// Ensure we start with a clean state
	cleanupDockerCompose(t)
	defer cleanupDockerCompose(t)

	// Start services
	t.Log("Starting Docker Compose services...")
	cmd := exec.Command("docker", "compose", "up", "-d", "--build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start services: %v\nOutput: %s", err, output)
	}

	// Wait for services to be healthy
	t.Log("Waiting for services to become healthy...")
	if !waitForServiceHealth(t, "http://localhost:8080/v1/health", 60*time.Second) {
		t.Fatal("API service failed to become healthy")
	}
	if !waitForServiceHealth(t, "http://localhost:3000/health", 60*time.Second) {
		t.Fatal("UI service failed to become healthy")
	}

	// Run regression tests
	t.Run("API_Service_Health", testAPIServiceHealth)
	t.Run("UI_Service_Health", testUIServiceHealth)
	t.Run("API_Endpoints_Accessible", testAPIEndpointsAccessible)
	t.Run("UI_Pages_Accessible", testUIPagesAccessible)
	t.Run("Cookie_Configuration_Check", testCookieConfigurationCheck)
	t.Run("RSS_Feed_URL_Generation", testRSSFeedURLGeneration)
	t.Run("UI_Error_Handling", func(t *testing.T) {
		// Run the UI error handling test as part of regression suite
		TestUIErrorHandlingRegression(t)
	})
}

func testAPIServiceHealth(t *testing.T) {
	resp, err := http.Get("http://localhost:8080/v1/health")
	if err != nil {
		t.Fatalf("Failed to get API health: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	status, ok := health["status"].(string)
	if !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}

	// Check that all required health checks pass
	checks, ok := health["checks"].(map[string]any)
	if !ok {
		t.Fatal("Health response missing checks")
	}

	requiredChecks := []string{"database", "storage", "core_service"}
	for _, check := range requiredChecks {
		if checkStatus, exists := checks[check]; !exists || checkStatus != "healthy" {
			t.Errorf("Health check '%s' failed: %v", check, checkStatus)
		}
	}
}

func testUIServiceHealth(t *testing.T) {
	resp, err := http.Get("http://localhost:3000/health")
	if err != nil {
		t.Fatalf("Failed to get UI health: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var health map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		t.Fatalf("Failed to decode health response: %v", err)
	}

	status, ok := health["status"].(string)
	if !ok || status != "healthy" {
		t.Errorf("Expected status 'healthy', got %v", health["status"])
	}

	// Check that all required health checks pass
	checks, ok := health["checks"].(map[string]any)
	if !ok {
		t.Fatal("Health response missing checks")
	}

	requiredChecks := []string{"api_service", "templates"}
	for _, check := range requiredChecks {
		if checkStatus, exists := checks[check]; !exists || checkStatus != "healthy" {
			t.Errorf("Health check '%s' failed: %v", check, checkStatus)
		}
	}
}

func testAPIEndpointsAccessible(t *testing.T) {
	endpoints := []struct {
		path           string
		expectedStatus int
	}{
		{"/v1/health", http.StatusOK},
		{"/v1/feeds", http.StatusOK},
		{"/v1/items", http.StatusOK},
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.path, func(t *testing.T) {
			resp, err := http.Get("http://localhost:8080" + endpoint.path)
			if err != nil {
				t.Fatalf("Failed to get %s: %v", endpoint.path, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != endpoint.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", endpoint.expectedStatus, endpoint.path, resp.StatusCode)
			}
		})
	}
}

func testUIPagesAccessible(t *testing.T) {
	endpoints := []struct {
		path           string
		expectedStatus int
	}{
		{"/health", http.StatusOK},
		{"/index.html", http.StatusOK},
		{"/", http.StatusOK}, // May return 200 if redirect is followed automatically
	}

	for _, endpoint := range endpoints {
		t.Run(endpoint.path, func(t *testing.T) {
			resp, err := http.Get("http://localhost:3000" + endpoint.path)
			if err != nil {
				t.Fatalf("Failed to get %s: %v", endpoint.path, err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != endpoint.expectedStatus {
				t.Errorf("Expected status %d for %s, got %d", endpoint.expectedStatus, endpoint.path, resp.StatusCode)
			}
		})
	}
}

func testCookieConfigurationCheck(t *testing.T) {
	// Get API service logs to check cookie configuration messages
	cmd := exec.Command("docker", "compose", "logs", "api-service")
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("Failed to get API service logs: %v", err)
	}

	logs := string(output)

	// Check docker-compose.yml to see if cookie file is configured
	composeCmd := exec.Command("docker", "compose", "config")
	composeOutput, err := composeCmd.Output()
	if err != nil {
		t.Fatalf("Failed to get docker compose config: %v", err)
	}
	composeConfig := string(composeOutput)

	// Check that cookie configuration is properly detected and reported
	if strings.Contains(composeConfig, "YTDLP_COOKIES_FILE") {
		// If cookie file is configured in docker-compose, should see appropriate messages
		if !strings.Contains(logs, "Cookie file configured:") {
			t.Error("Cookie file configured in docker-compose but no configuration message found in logs")
		}
		// Should also see permission checks
		if !strings.Contains(logs, "Cookie file exists") && !strings.Contains(logs, "Cookie directory") {
			t.Error("Cookie file configured but no permission check messages found in logs")
		}
	} else {
		// If no cookie file configured, should see the optional message
		if !strings.Contains(logs, "No cookie file configured (optional feature not enabled)") {
			t.Error("No cookie file configured but missing optional feature message in logs")
		}
	}
}

func testRSSFeedURLGeneration(t *testing.T) {
	// First, add a test item to ensure we have data
	addTestItem(t)

	// Get the UI page
	resp, err := http.Get("http://localhost:3000/index.html")
	if err != nil {
		t.Fatalf("Failed to get UI page: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}

	// Read the response body
	body := make([]byte, 10000)
	n, err := resp.Body.Read(body)
	if err != nil && err.Error() != "EOF" {
		t.Fatalf("Failed to read response body: %v", err)
	}
	pageContent := string(body[:n])

	// Check that RSS feed URLs point to the correct API service port (8080), not UI port (3000)
	if strings.Contains(pageContent, "RSS Feed") {
		// If there are RSS feed links, they should point to localhost:8080, not localhost:3000
		if strings.Contains(pageContent, "localhost:3000/v1/feeds") {
			t.Error("RSS feed URLs incorrectly point to UI service (port 3000) instead of API service (port 8080)")
		}
		if !strings.Contains(pageContent, "localhost:8080/v1/feeds") {
			t.Error("RSS feed URLs should point to API service (localhost:8080/v1/feeds)")
		}
	}
}

func addTestItem(t *testing.T) {
	// Add a test YouTube video to ensure we have data for RSS feed testing
	testURL := "https://www.youtube.com/watch?v=jNQXAC9IVRw" // "Me at the zoo" - first YouTube video

	payload := fmt.Sprintf(`{"urls":["%s"]}`, testURL)
	resp, err := http.Post("http://localhost:8080/v1/addItems", "application/json", strings.NewReader(payload))
	if err != nil {
		t.Logf("Failed to add test item (this is expected if the item already exists): %v", err)
		return
	}
	defer func() { _ = resp.Body.Close() }()

	// Don't fail the test if adding the item fails - it might already exist
	if resp.StatusCode != http.StatusOK {
		t.Logf("Failed to add test item (status %d) - this is expected if the item already exists", resp.StatusCode)
	}

	// Wait a bit for processing
	time.Sleep(2 * time.Second)
}

func waitForServiceHealth(t *testing.T, healthURL string, timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			resp, err := http.Get(healthURL)
			if err != nil {
				t.Logf("Health check failed: %v", err)
				continue
			}
			_ = resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				return true
			}
			t.Logf("Service not healthy yet (status %d)", resp.StatusCode)
		}
	}
}

func cleanupDockerCompose(t *testing.T) {
	cmd := exec.Command("docker", "compose", "down", "-v")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Logf("Failed to cleanup docker compose (this is expected if services weren't running): %v\nOutput: %s", err, output)
	}
}

// TestUIErrorHandlingRegression ensures that UI properly displays error messages for API failures
func TestUIErrorHandlingRegression(t *testing.T) {
	// Skip if not in CI or if explicitly requested
	if os.Getenv("SKIP_REGRESSION_TESTS") == "true" {
		t.Skip("Regression tests skipped")
	}

	// This test ensures that when the API returns errors, the UI shows appropriate error messages
	// This prevents regression of the issue where 500 errors might not show error messages

	// Start services if not already running
	cleanupDockerCompose(t)
	defer cleanupDockerCompose(t)

	cmd := exec.Command("docker", "compose", "up", "-d", "--build")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to start services: %v\nOutput: %s", err, output)
	}

	// Wait for services to be healthy
	if !waitForServiceHealth(t, "http://localhost:8080/v1/health", 60*time.Second) {
		t.Fatal("API service failed to become healthy")
	}
	if !waitForServiceHealth(t, "http://localhost:3000/health", 60*time.Second) {
		t.Fatal("UI service failed to become healthy")
	}

	// Test different error scenarios
	testCases := []struct {
		name        string
		url         string
		expectError bool
		description string
	}{
		{
			name:        "Invalid_URL_500_Error",
			url:         "https://www.youtube.com/watch?v=invalid",
			expectError: true,
			description: "Should show error message for invalid video URL that causes 500 error",
		},
		{
			name:        "Malformed_URL_400_Error",
			url:         "not-a-url",
			expectError: true,
			description: "Should show error message for malformed URL",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Submit form data to HTMX endpoint
			payload := fmt.Sprintf("url=%s", tc.url)
			resp, err := http.Post("http://localhost:3000/htmx/addItem", "application/x-www-form-urlencoded", strings.NewReader(payload))
			if err != nil {
				t.Fatalf("Failed to submit form: %v", err)
			}
			defer func() { _ = resp.Body.Close() }()

			// Read response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("Failed to read response: %v", err)
			}
			responseBody := string(body)

			if tc.expectError {
				// Should return an error status
				if resp.StatusCode < 400 {
					t.Errorf("Expected error status (>=400), got %d", resp.StatusCode)
				}

				// Should contain an error message (not be empty)
				if responseBody == "" {
					t.Error("Response body is empty - no error message shown to user!")
				}

				// Should contain HTML span with error styling
				if !strings.Contains(responseBody, "<span") {
					t.Errorf("Expected HTML error message, got: %s", responseBody)
				}

				// Should contain error styling
				if !strings.Contains(responseBody, "color:red") && !strings.Contains(responseBody, "color:orange") {
					t.Errorf("Expected styled error message, got: %s", responseBody)
				}

				// Should contain user-friendly message (not technical details)
				if strings.Contains(responseBody, "HTTP 500") || strings.Contains(responseBody, "stack trace") {
					t.Errorf("Error message contains technical details that should be hidden from users: %s", responseBody)
				}

				t.Logf("✅ %s: Status %d, Message: %s", tc.description, resp.StatusCode, responseBody)
			} else {
				// Should return success status
				if resp.StatusCode >= 400 {
					t.Errorf("Expected success status (<400), got %d: %s", resp.StatusCode, responseBody)
				}
			}
		})
	}
}
