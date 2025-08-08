package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 100*time.Millisecond)

	// Test initial state is closed
	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected initial state to be closed, got %v", cb.GetState())
	}

	// Simulate failures to open the circuit
	for i := 0; i < 3; i++ {
		err := cb.Call(func() error {
			return fmt.Errorf("simulated failure")
		})
		if err == nil {
			t.Errorf("Expected error on failure %d", i+1)
		}
	}

	// Circuit should now be open
	if cb.GetState() != CircuitOpen {
		t.Errorf("Expected circuit to be open after failures, got %v", cb.GetState())
	}

	// Calls should be rejected while circuit is open
	err := cb.Call(func() error {
		return nil
	})
	if err == nil || !strings.Contains(err.Error(), "circuit breaker is open") {
		t.Errorf("Expected circuit breaker error, got %v", err)
	}

	// Wait for timeout and test half-open state
	time.Sleep(150 * time.Millisecond)

	// First call should transition to half-open
	err = cb.Call(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected successful call in half-open state, got %v", err)
	}

	// After successful calls, circuit should close
	for i := 0; i < 2; i++ {
		err = cb.Call(func() error {
			return nil
		})
		if err != nil {
			t.Errorf("Expected successful call %d in half-open state, got %v", i+1, err)
		}
	}

	// Circuit should now be closed
	if cb.GetState() != CircuitClosed {
		t.Errorf("Expected circuit to be closed after successful calls, got %v", cb.GetState())
	}
}

func TestAPIClientRetryLogic(t *testing.T) {
	// Create a test server that fails the first 2 requests, then succeeds
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Server error"))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("[]"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 5*time.Second)

	// This should succeed after retries
	_, err := client.GetAllPodcastItems()
	if err != nil {
		t.Errorf("Expected success after retries, got %v", err)
	}

	if callCount != 3 {
		t.Errorf("Expected 3 calls (2 failures + 1 success), got %d", callCount)
	}
}

func TestAPIClientCircuitBreakerIntegration(t *testing.T) {
	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Server error"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 1*time.Second).(*HTTPAPIClient)

	// Make enough calls to open the circuit breaker
	for i := 0; i < 6; i++ {
		_, err := client.GetAllPodcastItems()
		if err == nil {
			t.Errorf("Expected error on call %d", i+1)
		}
	}

	// Circuit should now be open, next call should fail immediately
	start := time.Now()
	_, err := client.GetAllPodcastItems()
	duration := time.Since(start)

	if err == nil {
		t.Errorf("Expected circuit breaker error")
	}

	if !strings.Contains(err.Error(), "circuit breaker") {
		t.Errorf("Expected circuit breaker error, got %v", err)
	}

	// Should fail quickly (no retries when circuit is open)
	if duration > 500*time.Millisecond {
		t.Errorf("Expected quick failure when circuit is open, took %v", duration)
	}
}

func TestAPIClientGracefulDegradation(t *testing.T) {
	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Server error"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 1*time.Second).(*HTTPAPIClient)

	// Open the circuit breaker first
	for i := 0; i < 6; i++ {
		_, _ = client.GetAllPodcastItems()
	}

	// Now test graceful degradation
	items, err := client.GetAllPodcastItems()

	// Should return empty slice with graceful degradation error
	if items == nil {
		t.Errorf("Expected empty slice, got nil")
	}

	if len(items) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(items))
	}

	if err == nil {
		t.Errorf("Expected graceful degradation error")
	}

	if !strings.Contains(err.Error(), "graceful degradation") {
		t.Errorf("Expected graceful degradation error, got %v", err)
	}
}

func TestAPIClientAddItemsSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/v1/addItems" {
			t.Errorf("Expected /v1/addItems path, got %s", r.URL.Path)
		}

		var requestBody map[string][]string
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		urls, ok := requestBody["urls"]
		if !ok || len(urls) != 1 || urls[0] != "https://example.com/video" {
			t.Errorf("Expected single URL, got %v", urls)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 5*time.Second)

	err := client.AddItems([]string{"https://example.com/video"})
	if err != nil {
		t.Errorf("Expected success, got %v", err)
	}
}

func TestAPIClientDeletePodcastItemSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		expectedPath := "/v1/feeds/testfeed/testitem"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected %s path, got %s", expectedPath, r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 5*time.Second)

	err := client.DeletePodcastItem("testfeed", "testitem")
	if err != nil {
		t.Errorf("Expected success, got %v", err)
	}
}

func TestAPIClientHealthCheckSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/health" {
			t.Errorf("Expected /v1/health path, got %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 5*time.Second)

	err := client.HealthCheck()
	if err != nil {
		t.Errorf("Expected success, got %v", err)
	}
}

func TestAPIClientGetFeedsWithGracefulDegradation(t *testing.T) {
	// Create a test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Server error"))
	}))
	defer server.Close()

	client := NewAPIClient(server.URL, 1*time.Second).(*HTTPAPIClient)

	// Open the circuit breaker first
	for i := 0; i < 6; i++ {
		_, _ = client.GetFeeds()
	}

	// Now test graceful degradation
	feeds, err := client.GetFeeds()

	// Should return empty slice with graceful degradation error
	if feeds == nil {
		t.Errorf("Expected empty slice, got nil")
	}

	if len(feeds) != 0 {
		t.Errorf("Expected empty slice, got %d feeds", len(feeds))
	}

	if err == nil {
		t.Errorf("Expected graceful degradation error")
	}

	if !strings.Contains(err.Error(), "graceful degradation") {
		t.Errorf("Expected graceful degradation error, got %v", err)
	}
}

func TestHTTPErrorHandling(t *testing.T) {
	// Test 4xx errors are not retried (except 408 and 429)
	testCases := []struct {
		statusCode  int
		shouldRetry bool
		description string
	}{
		{400, false, "Bad Request should not retry"},
		{401, false, "Unauthorized should not retry"},
		{404, false, "Not Found should not retry"},
		{408, true, "Request Timeout should retry"},
		{429, true, "Too Many Requests should retry"},
		{500, true, "Internal Server Error should retry"},
		{502, true, "Bad Gateway should retry"},
		{503, true, "Service Unavailable should retry"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			callCount := 0
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.WriteHeader(tc.statusCode)
				_, _ = w.Write([]byte("Error"))
			}))
			defer server.Close()

			client := NewAPIClient(server.URL, 1*time.Second)

			_, err := client.GetAllPodcastItems()

			if err == nil {
				t.Errorf("Expected error for status %d", tc.statusCode)
			}

			expectedCalls := 1
			if tc.shouldRetry {
				expectedCalls = 4 // 1 initial + 3 retries
			}

			if callCount != expectedCalls {
				t.Errorf("Expected %d calls for status %d, got %d", expectedCalls, tc.statusCode, callCount)
			}
		})
	}
}
