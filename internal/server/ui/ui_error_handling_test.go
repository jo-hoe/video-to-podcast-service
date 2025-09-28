package ui

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
)

// MockAPIClient for testing error scenarios
type MockAPIClient struct {
	shouldReturnError bool
	errorToReturn     error
}

func (m *MockAPIClient) AddItems(urls []string) error {
	if m.shouldReturnError {
		return m.errorToReturn
	}
	return nil
}

func (m *MockAPIClient) GetAllPodcastItems() ([]*database.PodcastItem, error) {
	return []*database.PodcastItem{}, nil
}

func (m *MockAPIClient) GetFeeds() ([]string, error) {
	return []string{}, nil
}

func (m *MockAPIClient) DeletePodcastItem(feedTitle, podcastItemID string) error {
	return nil
}

func (m *MockAPIClient) GetLinkToFeed(host, feedsPath, filePath string) string {
	return "http://localhost:8080/v1/feeds/test/rss.xml"
}

func (m *MockAPIClient) HealthCheck() error {
	return nil
}

// HTTPError represents an HTTP error with status code (same as in apiclient.go)
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// TestUIErrorHandling tests that the UI properly displays error messages for different HTTP status codes
func TestUIErrorHandling(t *testing.T) {
	tests := []struct {
		name           string
		error          error
		expectedStatus int
		expectedBody   string
		description    string
	}{
		{
			name:           "HTTP_500_Error",
			error:          &HTTPError{StatusCode: 500, Message: `{"message":"failed to process download item"}`},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Server error while processing the video. Please try again later.",
			description:    "Should show server error message for HTTP 500",
		},
		{
			name:           "HTTP_400_Error",
			error:          &HTTPError{StatusCode: 400, Message: `{"message":"invalid URL"}`},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid URL format or unsupported video source.",
			description:    "Should show invalid URL message for HTTP 400",
		},
		{
			name:           "Circuit_Breaker_Error",
			error:          fmt.Errorf("circuit breaker is open"),
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "Processing service is temporarily unavailable. Please try again in a few minutes.",
			description:    "Should show circuit breaker message",
		},
		{
			name:           "Connection_Error",
			error:          fmt.Errorf("failed to send request to API service: connection refused"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Unable to connect to the processing service. Please try again later.",
			description:    "Should show connection error message",
		},
		{
			name:           "Retry_Failure_Error",
			error:          fmt.Errorf("operation failed after 3 retries: HTTP 500: {\"message\":\"failed to process download item\"}"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Server error while processing the video. Please try again later.",
			description:    "Should detect HTTP 500 in retry failure message",
		},
		{
			name:           "Generic_Error",
			error:          fmt.Errorf("some unknown error"),
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "Failed to process video. Please check the URL and try again.",
			description:    "Should show generic error message for unknown errors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock API client
			mockClient := &MockAPIClient{
				shouldReturnError: true,
				errorToReturn:     tt.error,
			}

			// Create UI service with mock client
			uiService := NewUIService(mockClient)

			// Create Echo instance and set up routes
			e := echo.New()
			uiService.SetUIRoutes(e)

			// Create test request
			req := httptest.NewRequest(http.MethodPost, "/htmx/addItem", strings.NewReader("url=https://www.youtube.com/watch?v=test"))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
			rec := httptest.NewRecorder()

			// Execute request through the HTTP server
			e.ServeHTTP(rec, req)

			// Check status code
			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			// Check response body contains expected message
			body := rec.Body.String()
			if !strings.Contains(body, tt.expectedBody) {
				t.Errorf("Expected body to contain '%s', got '%s'\nDescription: %s", tt.expectedBody, body, tt.description)
			}
		})
	}
}

// TestUIErrorHandlingUnit ensures that 500 errors are properly displayed
func TestUIErrorHandlingUnit(t *testing.T) {
	// This test specifically addresses the issue where 500 errors might not show error messages

	// Test the exact error format that comes from the API client
	retryError := fmt.Errorf("operation failed after 3 retries: HTTP 500: {\"message\":\"failed to process download item\"}")

	mockClient := &MockAPIClient{
		shouldReturnError: true,
		errorToReturn:     retryError,
	}

	uiService := NewUIService(mockClient)
	e := echo.New()
	uiService.SetUIRoutes(e)

	req := httptest.NewRequest(http.MethodPost, "/htmx/addItem", strings.NewReader("url=https://www.youtube.com/watch?v=test"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()

	// Execute request through the HTTP server
	e.ServeHTTP(rec, req)

	// Should return 500 status
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", rec.Code)
	}

	// Should contain error message, not be empty
	body := rec.Body.String()
	if body == "" {
		t.Error("Response body is empty - no error message shown to user!")
	}

	// Should contain a user-friendly error message
	if !strings.Contains(body, "Server error while processing the video") {
		t.Errorf("Expected user-friendly error message, got: %s", body)
	}

	// Should be styled HTML for HTMX
	if !strings.Contains(body, "<span") || !strings.Contains(body, "color:red") {
		t.Errorf("Expected styled HTML error message, got: %s", body)
	}
}
