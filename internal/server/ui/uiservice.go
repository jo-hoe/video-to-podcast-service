package ui

import (
	"fmt"
	"net/http"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
)

const MainPageName = "index.html"
const FeedsPath = "v1/feeds" // API feeds path constant

// APIClient interface for communicating with the API service
type APIClient interface {
	AddItems(urls []string) error
	GetAllPodcastItems() ([]*database.PodcastItem, error)
	GetFeeds() ([]string, error)
	DeletePodcastItem(feedTitle, podcastItemID string) error
	GetLinkToFeed(host, feedsPath, filePath string) string
	HealthCheck() error
}

type UIService struct {
	apiClient APIClient
}

type PodcastItemList struct {
	PodcastItems []*database.PodcastItem
	Host         string
	APIPath      string
}

func NewUIService(apiClient APIClient) *UIService {
	return &UIService{
		apiClient: apiClient,
	}
}

func (service *UIService) SetUIRoutes(e *echo.Echo) {
	// Create template with helper functions
	funcMap := template.FuncMap{
		"formatDuration":       formatDuration,
		"getFeedLink":          service.getFeedLink,
		"getFeedTitleFromPath": getFeedTitleFromPath,
	}

	e.Renderer = &Template{
		templates: template.Must(template.New("").Funcs(funcMap).ParseFS(templateFS, viewsPattern)),
	}

	// Set UI routes
	e.GET(MainPageName, service.indexHandler)
	// redirect to main page
	e.GET("/", func(c echo.Context) error {
		return c.Redirect(http.StatusFound, "/"+MainPageName)
	})

	e.POST("/htmx/addItem", service.htmxAddItemHandler)
	e.DELETE("/htmx/deleteItem/:feedTitle/:podcastItemID", service.htmxDeleteItemHandler)

	// Health check endpoint
	e.GET("/health", service.healthHandler)
}

// Helper function to extract feed title from audio file path
func getFeedTitleFromPath(path string) string {
	// Assumes feed title is the parent directory of the audio file
	return filepath.Base(filepath.Dir(path))
}

func (service *UIService) indexHandler(ctx echo.Context) (err error) {
	// Get all podcast items from API service
	podcastItems, err := service.apiClient.GetAllPodcastItems()
	if err != nil {
		// Check if this is a graceful degradation error
		if isGracefulDegradationError(err) {
			ctx.Logger().Warnf("API service unavailable, showing cached/empty data: %v", err)
			podcastItems = []*database.PodcastItem{}
		} else {
			// Log error but don't fail the page load - show empty list with error indication
			ctx.Logger().Errorf("Failed to fetch podcast items from API service: %v", err)
			podcastItems = []*database.PodcastItem{}
		}
	}

	// Sort by UpdatedAt in descending order (most recent first)
	sort.Slice(podcastItems, func(i, j int) bool {
		return podcastItems[i].UpdatedAt.After(podcastItems[j].UpdatedAt)
	})

	// reduce the number of results to 128 items
	if len(podcastItems) > 128 {
		podcastItems = podcastItems[:128]
	}

	// Get host and API path from the request
	// For RSS feed links, we need to use the external API URL, not the UI host
	// Convert internal API URL to external URL by replacing the service name with localhost
	// and using the external port (8080)
	host := "localhost:8080" // External API service URL
	apiPath := "api"         // You might want to make this configurable

	data := PodcastItemList{
		PodcastItems: podcastItems,
		Host:         host,
		APIPath:      apiPath,
	}

	return ctx.Render(http.StatusOK, "index", data)
}

// New handler for HTMX single URL form
func (service *UIService) htmxAddItemHandler(ctx echo.Context) error {
	type SingleUrl struct {
		URL string `json:"url" form:"url" validate:"required"`
	}
	var req SingleUrl
	if err := ctx.Bind(&req); err != nil || req.URL == "" {
		return ctx.HTML(http.StatusBadRequest, "<span style='color:red'>Invalid or missing URL.</span>")
	}

	if err := service.apiClient.AddItems([]string{req.URL}); err != nil {
		ctx.Logger().Errorf("Failed to add item via API client: %v", err)

		// Check for circuit breaker errors first
		if strings.Contains(err.Error(), "circuit breaker") {
			return ctx.HTML(http.StatusServiceUnavailable, "<span style='color:orange'>Processing service is temporarily unavailable. Please try again in a few minutes.</span>")
		}

		// Provide more specific error messages based on error content
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "HTTP 400") {
			return ctx.HTML(http.StatusBadRequest, "<span style='color:red'>Invalid URL format or unsupported video source.</span>")
		} else if strings.Contains(errorMsg, "HTTP 500") {
			return ctx.HTML(http.StatusInternalServerError, "<span style='color:red'>Server error while processing the video. Please try again later.</span>")
		} else if strings.Contains(errorMsg, "failed to send request") || strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
			return ctx.HTML(http.StatusInternalServerError, "<span style='color:red'>Unable to connect to the processing service. Please try again later.</span>")
		} else if strings.Contains(errorMsg, "operation failed after") && strings.Contains(errorMsg, "retries") {
			return ctx.HTML(http.StatusInternalServerError, "<span style='color:orange'>Processing service is experiencing issues. Please try again in a few minutes.</span>")
		}

		// Generic error for other problems
		return ctx.HTML(http.StatusInternalServerError, "<span style='color:red'>Failed to process video. Please check the URL and try again.</span>")
	}
	return ctx.HTML(http.StatusOK, "<span style='color:green'>Video submitted successfully! Processing will begin shortly.</span>")
}

// Helper function to format duration from milliseconds to HH:MM:SS or MM:SS
func formatDuration(milliseconds int64) string {
	totalSeconds := milliseconds / 1000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, seconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, seconds)
}

// HTMX handler for deleting podcast items
func (service *UIService) htmxDeleteItemHandler(ctx echo.Context) error {
	feedTitle := ctx.Param("feedTitle")
	podcastItemID := ctx.Param("podcastItemID")

	if feedTitle == "" || podcastItemID == "" {
		return ctx.HTML(http.StatusBadRequest, "<span style='color:red'>Missing feed title or podcast item ID.</span>")
	}

	err := service.apiClient.DeletePodcastItem(feedTitle, podcastItemID)
	if err != nil {
		ctx.Logger().Errorf("Failed to delete item via API client: %v", err)

		// Check for circuit breaker errors first
		if strings.Contains(err.Error(), "circuit breaker") {
			return ctx.HTML(http.StatusServiceUnavailable, "<span style='color:orange'>Service is temporarily unavailable. Please try again in a few minutes.</span>")
		}

		// Provide more specific error messages based on error content
		errorMsg := err.Error()
		if strings.Contains(errorMsg, "HTTP 404") {
			return ctx.HTML(http.StatusNotFound, "<span style='color:red'>Item not found or already deleted.</span>")
		} else if strings.Contains(errorMsg, "HTTP 400") {
			return ctx.HTML(http.StatusBadRequest, "<span style='color:red'>Invalid item or feed information.</span>")
		} else if strings.Contains(errorMsg, "failed to send request") || strings.Contains(errorMsg, "connection") || strings.Contains(errorMsg, "timeout") {
			return ctx.HTML(http.StatusInternalServerError, "<span style='color:red'>Unable to connect to the service. Please try again later.</span>")
		} else if strings.Contains(errorMsg, "operation failed after") && strings.Contains(errorMsg, "retries") {
			return ctx.HTML(http.StatusInternalServerError, "<span style='color:orange'>Service is experiencing issues. Please try again in a few minutes.</span>")
		}

		return ctx.HTML(http.StatusInternalServerError, "<span style='color:red'>Failed to delete item. Please try again.</span>")
	}

	// Return empty content to remove the element from the DOM
	return ctx.NoContent(http.StatusOK)
}

// Helper function to generate feed link for a podcast item
func (service *UIService) getFeedLink(host, filePath string) string {
	return service.apiClient.GetLinkToFeed(host, FeedsPath, filePath)
}

// isGracefulDegradationError checks if an error is a graceful degradation error
func isGracefulDegradationError(err error) bool {
	// Check if the error message contains graceful degradation indicators
	errorMsg := err.Error()
	return strings.Contains(errorMsg, "graceful degradation") ||
		strings.Contains(errorMsg, "circuit breaker")
}

// UIHealthStatus represents the health status of the UI service
type UIHealthStatus struct {
	Status      string            `json:"status"`
	Timestamp   string            `json:"timestamp"`
	Version     string            `json:"version"`
	Uptime      string            `json:"uptime"`
	Checks      map[string]string `json:"checks"`
	ServiceInfo UIServiceInfo     `json:"service_info"`
}

type UIServiceInfo struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Port        string `json:"port"`
}

var uiServiceStartTime = time.Now()

// healthHandler provides comprehensive health check for the UI service
func (service *UIService) healthHandler(ctx echo.Context) error {
	// Perform health checks
	checks := make(map[string]string)
	overallStatus := "healthy"

	// Check API service connectivity
	if apiErr := service.checkAPIServiceHealth(); apiErr != nil {
		checks["api_service"] = fmt.Sprintf("unhealthy: %v", apiErr)
		overallStatus = "unhealthy"
	} else {
		checks["api_service"] = "healthy"
	}

	// Check template rendering system
	if templateErr := service.checkTemplateHealth(); templateErr != nil {
		checks["templates"] = fmt.Sprintf("unhealthy: %v", templateErr)
		overallStatus = "unhealthy"
	} else {
		checks["templates"] = "healthy"
	}

	uptime := time.Since(uiServiceStartTime)

	healthStatus := UIHealthStatus{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0", // You might want to make this configurable
		Uptime:    uptime.String(),
		Checks:    checks,
		ServiceInfo: UIServiceInfo{
			Name:        "video-to-podcast-ui",
			Description: "UI service for video to podcast conversion",
			Port:        "3000", // You might want to make this configurable
		},
	}

	// Return appropriate HTTP status based on health
	if overallStatus == "healthy" {
		return ctx.JSON(http.StatusOK, healthStatus)
	} else {
		return ctx.JSON(http.StatusServiceUnavailable, healthStatus)
	}
}

// checkAPIServiceHealth verifies API service connectivity
func (service *UIService) checkAPIServiceHealth() error {
	// Use the existing health check method from the API client
	return service.apiClient.HealthCheck()
}

// checkTemplateHealth verifies template system functionality
func (service *UIService) checkTemplateHealth() error {
	// Basic check to ensure templates are loaded
	// This is a simple validation that the template system is working
	if service == nil {
		return fmt.Errorf("UI service is nil")
	}

	// We could add more sophisticated template checks here if needed
	// For now, just verify the service is properly initialized
	if service.apiClient == nil {
		return fmt.Errorf("API client not initialized")
	}

	return nil
}
