package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
)

// APIClient interface for communicating with the API service
type APIClient interface {
	AddItems(urls []string) error
	GetAllPodcastItems() ([]*database.PodcastItem, error)
	GetFeeds() ([]string, error)
	DeletePodcastItem(feedTitle, podcastItemID string) error
	GetLinkToFeed(host, feedsPath, filePath string) string
	HealthCheck() error
}

// CircuitBreakerState represents the state of the circuit breaker
type CircuitBreakerState int

const (
	CircuitClosed CircuitBreakerState = iota
	CircuitOpen
	CircuitHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu               sync.RWMutex
	state            CircuitBreakerState
	failureCount     int
	lastFailureTime  time.Time
	successCount     int
	maxFailures      int
	timeout          time.Duration
	halfOpenMaxCalls int
}

// NewCircuitBreaker creates a new circuit breaker
func NewCircuitBreaker(maxFailures int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            CircuitClosed,
		maxFailures:      maxFailures,
		timeout:          timeout,
		halfOpenMaxCalls: 3,
	}
}

// Call executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// Check if circuit should transition from open to half-open
	if cb.state == CircuitOpen && time.Since(cb.lastFailureTime) > cb.timeout {
		cb.state = CircuitHalfOpen
		cb.successCount = 0
		log.Printf("[CircuitBreaker] Transitioning to half-open state")
	}

	// Reject calls if circuit is open
	if cb.state == CircuitOpen {
		return &CircuitBreakerError{Message: "circuit breaker is open"}
	}

	// Limit calls in half-open state
	if cb.state == CircuitHalfOpen && cb.successCount >= cb.halfOpenMaxCalls {
		return &CircuitBreakerError{Message: "circuit breaker is half-open, max calls exceeded"}
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

func (cb *CircuitBreaker) onSuccess() {
	cb.failureCount = 0

	if cb.state == CircuitHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenMaxCalls {
			cb.state = CircuitClosed
			log.Printf("[CircuitBreaker] Transitioning to closed state after successful calls")
		}
	}
}

func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.state == CircuitClosed && cb.failureCount >= cb.maxFailures {
		cb.state = CircuitOpen
		log.Printf("[CircuitBreaker] Transitioning to open state after %d failures", cb.failureCount)
	} else if cb.state == CircuitHalfOpen {
		cb.state = CircuitOpen
		log.Printf("[CircuitBreaker] Transitioning back to open state from half-open")
	}
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// CircuitBreakerError represents an error from the circuit breaker
type CircuitBreakerError struct {
	Message string
}

func (e *CircuitBreakerError) Error() string {
	return fmt.Sprintf("circuit breaker error: %s", e.Message)
}

// HTTPAPIClient implements APIClient using HTTP requests with comprehensive error handling
type HTTPAPIClient struct {
	baseURL        string
	httpClient     *http.Client
	maxRetries     int
	circuitBreaker *CircuitBreaker
	logger         *log.Logger
}

// NewAPIClient creates a new API client with comprehensive error handling
func NewAPIClient(baseURL string, timeout time.Duration) APIClient {
	return &HTTPAPIClient{
		baseURL:    baseURL,
		maxRetries: 3,
		httpClient: &http.Client{
			Timeout: timeout,
		},
		circuitBreaker: NewCircuitBreaker(5, 30*time.Second), // Open after 5 failures, retry after 30s
		logger:         log.New(log.Writer(), "[APIClient] ", log.LstdFlags|log.Lshortfile),
	}
}

// closeResponseBody safely closes the response body and logs any errors
func (c *HTTPAPIClient) closeResponseBody(resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		c.logger.Printf("Warning: failed to close response body: %v", err)
	}
}

// AddItems sends URLs to the API service for processing
func (c *HTTPAPIClient) AddItems(urls []string) error {
	requestBody := map[string][]string{
		"urls": urls,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		c.logger.Printf("Failed to marshal AddItems request: %v", err)
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	return c.executeWithCircuitBreaker("AddItems", func() error {
		return c.retryWithBackoff("AddItems", func() error {
			resp, err := c.httpClient.Post(
				c.baseURL+"/v1/addItems",
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				c.logger.Printf("AddItems request failed: %v", err)
				return fmt.Errorf("failed to send request to API service: %w", err)
			}
			defer c.closeResponseBody(resp)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				c.logger.Printf("AddItems returned HTTP %d: %s", resp.StatusCode, string(body))
				return &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    string(body),
				}
			}

			c.logger.Printf("AddItems completed successfully for %d URLs", len(urls))
			return nil
		})
	})
}

// GetAllPodcastItems retrieves all podcast items from the API service with graceful degradation
func (c *HTTPAPIClient) GetAllPodcastItems() ([]*database.PodcastItem, error) {
	var items []*database.PodcastItem

	err := c.executeWithCircuitBreaker("GetAllPodcastItems", func() error {
		return c.retryWithBackoff("GetAllPodcastItems", func() error {
			resp, err := c.httpClient.Get(c.baseURL + "/v1/items")
			if err != nil {
				c.logger.Printf("GetAllPodcastItems request failed: %v", err)
				return fmt.Errorf("failed to get podcast items from API service: %w", err)
			}
			defer c.closeResponseBody(resp)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				c.logger.Printf("GetAllPodcastItems returned HTTP %d: %s", resp.StatusCode, string(body))
				return &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    string(body),
				}
			}

			if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
				c.logger.Printf("Failed to decode GetAllPodcastItems response: %v", err)
				return fmt.Errorf("failed to decode response: %w", err)
			}

			c.logger.Printf("GetAllPodcastItems completed successfully, retrieved %d items", len(items))
			return nil
		})
	})

	// Graceful degradation: return empty slice if API is unavailable
	if err != nil {
		c.logger.Printf("GetAllPodcastItems failed with graceful degradation: %v", err)
		if c.isCircuitBreakerError(err) {
			c.logger.Printf("Returning empty podcast items due to circuit breaker")
			return []*database.PodcastItem{}, &GracefulDegradationError{
				Operation: "GetAllPodcastItems",
				Cause:     err,
			}
		}
		return nil, err
	}

	return items, nil
}

// GetFeeds retrieves all available feeds from the API service with graceful degradation
func (c *HTTPAPIClient) GetFeeds() ([]string, error) {
	var feeds []string

	err := c.executeWithCircuitBreaker("GetFeeds", func() error {
		return c.retryWithBackoff("GetFeeds", func() error {
			resp, err := c.httpClient.Get(c.baseURL + "/v1/feeds")
			if err != nil {
				c.logger.Printf("GetFeeds request failed: %v", err)
				return fmt.Errorf("failed to get feeds from API service: %w", err)
			}
			defer c.closeResponseBody(resp)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				c.logger.Printf("GetFeeds returned HTTP %d: %s", resp.StatusCode, string(body))
				return &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    string(body),
				}
			}

			if err := json.NewDecoder(resp.Body).Decode(&feeds); err != nil {
				c.logger.Printf("Failed to decode GetFeeds response: %v", err)
				return fmt.Errorf("failed to decode feeds response: %w", err)
			}

			c.logger.Printf("GetFeeds completed successfully, retrieved %d feeds", len(feeds))
			return nil
		})
	})

	// Graceful degradation: return empty slice if API is unavailable
	if err != nil {
		c.logger.Printf("GetFeeds failed with graceful degradation: %v", err)
		if c.isCircuitBreakerError(err) {
			c.logger.Printf("Returning empty feeds due to circuit breaker")
			return []string{}, &GracefulDegradationError{
				Operation: "GetFeeds",
				Cause:     err,
			}
		}
		return nil, err
	}

	return feeds, nil
}

// DeletePodcastItem deletes a podcast item from the API service
func (c *HTTPAPIClient) DeletePodcastItem(feedTitle, podcastItemID string) error {
	return c.executeWithCircuitBreaker("DeletePodcastItem", func() error {
		return c.retryWithBackoff("DeletePodcastItem", func() error {
			url := fmt.Sprintf("%s/v1/feeds/%s/%s", c.baseURL, feedTitle, podcastItemID)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			if err != nil {
				c.logger.Printf("Failed to create DeletePodcastItem request: %v", err)
				return fmt.Errorf("failed to create delete request: %w", err)
			}

			resp, err := c.httpClient.Do(req)
			if err != nil {
				c.logger.Printf("DeletePodcastItem request failed: %v", err)
				return fmt.Errorf("failed to delete podcast item: %w", err)
			}
			defer c.closeResponseBody(resp)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				c.logger.Printf("DeletePodcastItem returned HTTP %d: %s", resp.StatusCode, string(body))
				return &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    string(body),
				}
			}

			c.logger.Printf("DeletePodcastItem completed successfully for %s/%s", feedTitle, podcastItemID)
			return nil
		})
	})
}

// HealthCheck checks if the API service is healthy
func (c *HTTPAPIClient) HealthCheck() error {
	return c.executeWithCircuitBreaker("HealthCheck", func() error {
		return c.retryWithBackoff("HealthCheck", func() error {
			resp, err := c.httpClient.Get(c.baseURL + "/v1/health")
			if err != nil {
				c.logger.Printf("HealthCheck request failed: %v", err)
				return fmt.Errorf("failed to check API service health: %w", err)
			}
			defer c.closeResponseBody(resp)

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				c.logger.Printf("HealthCheck returned HTTP %d: %s", resp.StatusCode, string(body))
				return &HTTPError{
					StatusCode: resp.StatusCode,
					Message:    string(body),
				}
			}

			c.logger.Printf("HealthCheck completed successfully")
			return nil
		})
	})
}

// GetLinkToFeed generates a link to the RSS feed
func (c *HTTPAPIClient) GetLinkToFeed(host, feedsPath, filePath string) string {
	// Extract feed title from the file path (parent directory name)
	feedTitle := filepath.Base(filepath.Dir(filePath))
	// Construct the RSS feed URL using the provided host (external URL)
	return fmt.Sprintf("http://%s/%s/%s/rss.xml", host, feedsPath, feedTitle)
}

// executeWithCircuitBreaker wraps operations with circuit breaker protection
func (c *HTTPAPIClient) executeWithCircuitBreaker(operation string, fn func() error) error {
	return c.circuitBreaker.Call(func() error {
		err := fn()
		if err != nil {
			c.logger.Printf("Operation %s failed: %v", operation, err)
		}
		return err
	})
}

// retryWithBackoff executes a function with exponential backoff retry logic
func (c *HTTPAPIClient) retryWithBackoff(operation string, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 1s, 2s, 4s, 8s...
			backoffDuration := time.Duration(1<<uint(attempt-1)) * time.Second
			c.logger.Printf("Retrying %s (attempt %d/%d) after %v", operation, attempt+1, c.maxRetries+1, backoffDuration)
			time.Sleep(backoffDuration)
		}

		lastErr = fn()
		if lastErr == nil {
			if attempt > 0 {
				c.logger.Printf("Operation %s succeeded after %d retries", operation, attempt)
			}
			return nil
		}

		// Don't retry on client errors (4xx) except for 408 (timeout) and 429 (rate limit)
		if httpErr, ok := lastErr.(*HTTPError); ok {
			if httpErr.StatusCode >= 400 && httpErr.StatusCode < 500 &&
				httpErr.StatusCode != 408 && httpErr.StatusCode != 429 {
				c.logger.Printf("Not retrying %s due to client error: HTTP %d", operation, httpErr.StatusCode)
				return lastErr
			}
		}
	}

	c.logger.Printf("Operation %s failed after %d retries: %v", operation, c.maxRetries, lastErr)
	return fmt.Errorf("operation failed after %d retries: %w", c.maxRetries, lastErr)
}

// isCircuitBreakerError checks if an error is from the circuit breaker
func (c *HTTPAPIClient) isCircuitBreakerError(err error) bool {
	_, ok := err.(*CircuitBreakerError)
	return ok
}

// HTTPError represents an HTTP error with status code
type HTTPError struct {
	StatusCode int
	Message    string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Message)
}

// GracefulDegradationError represents an error that was handled with graceful degradation
type GracefulDegradationError struct {
	Operation string
	Cause     error
}

func (e *GracefulDegradationError) Error() string {
	return fmt.Sprintf("graceful degradation for %s: %v", e.Operation, e.Cause)
}

func (e *GracefulDegradationError) Unwrap() error {
	return e.Cause
}
