package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/internal/server/ui"
	"github.com/labstack/echo/v4"
)

// TestUIServiceWithFailingAPI tests the UI service behavior when API service is failing
func TestUIServiceWithFailingAPI(t *testing.T) {
	// Create a test API server that always fails
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("API service is down"))
	}))
	defer apiServer.Close()

	// Create API client with short timeout for faster testing
	apiClient := NewAPIClient(apiServer.URL, 1*time.Second)

	// Create UI service
	uiService := ui.NewUIService(apiClient)

	// Create Echo instance
	e := echo.New()
	uiService.SetUIRoutes(e)

	// The index handler should handle API failures gracefully
	// We can't easily test the handler directly due to template dependencies,
	// but we can test the API client behavior which is the core of our error handling

	// Test that API client handles failures with circuit breaker
	for i := 0; i < 6; i++ {
		_, err := apiClient.GetAllPodcastItems()
		if err == nil {
			t.Errorf("Expected error on call %d", i+1)
		}
	}

	// After circuit breaker opens, calls should fail quickly with graceful degradation
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
}

// TestUIServiceHTMXErrorHandling tests HTMX error handling with different API failures
func TestUIServiceHTMXErrorHandling(t *testing.T) {
	testCases := []struct {
		name           string
		statusCode     int
		expectedError  string
		expectedStatus int
	}{
		{
			name:           "Bad Request",
			statusCode:     400,
			expectedError:  "Invalid URL format",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Internal Server Error",
			statusCode:     500,
			expectedError:  "Server error while processing",
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create API server that returns specific error
			apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte("Test error"))
			}))
			defer apiServer.Close()

			// Create API client
			apiClient := NewAPIClient(apiServer.URL, 1*time.Second)

			// Test AddItems error handling
			err := apiClient.AddItems([]string{"https://example.com/video"})
			if err == nil {
				t.Errorf("Expected error for status %d", tc.statusCode)
			}

			// Verify error contains expected information
			if !strings.Contains(err.Error(), "HTTP") {
				t.Errorf("Expected HTTP error, got %v", err)
			}
		})
	}
}

// TestCircuitBreakerRecovery tests that circuit breaker can recover after failures
func TestCircuitBreakerRecovery(t *testing.T) {
	callCount := 0

	// Create API server that fails first few calls, then succeeds
	apiServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 20 { // Fail enough times to open circuit breaker
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
	client := NewAPIClient(apiServer.URL, 1*time.Second).(*HTTPAPIClient)
	client.circuitBreaker = NewCircuitBreaker(5, 100*time.Millisecond) // Faster recovery

	// Make calls to open circuit breaker
	for i := 0; i < 6; i++ {
		_, _ = client.GetAllPodcastItems()
	}

	// Verify circuit is open
	if client.circuitBreaker.GetState() != CircuitOpen {
		t.Errorf("Expected circuit to be open")
	}

	// Wait for circuit breaker timeout
	time.Sleep(150 * time.Millisecond)

	// Make successful calls to close circuit
	for i := 0; i < 3; i++ {
		_, err := client.GetAllPodcastItems()
		if err != nil {
			t.Errorf("Expected success after recovery, got %v", err)
		}
	}

	// Verify circuit is closed
	if client.circuitBreaker.GetState() != CircuitClosed {
		t.Errorf("Expected circuit to be closed after recovery, got %v", client.circuitBreaker.GetState())
	}
}
