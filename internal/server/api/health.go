package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
)

const (
	HealthPath = "health"
)

type HealthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks"`
	Message string            `json:"message,omitempty"`
}

// healthHandler provides a health check for Kubernetes readiness/liveness probes
func (service *APIService) healthHandler(ctx echo.Context) error {
	checks := map[string]string{
		"service":         "healthy",
		"cookies":         service.checkCookieHealth(),
		"database":        service.checkDatabaseHealth(),
		"media_directory": service.checkMediaHealth(),
	}

	// Check if any health check failed
	allHealthy := true
	var failures []string
	for name, status := range checks {
		if status != "healthy" && status != "disabled" {
			allHealthy = false
			failures = append(failures, name)
		}
	}

	response := HealthResponse{Checks: checks}
	if allHealthy {
		response.Status = "healthy"
		return ctx.JSON(http.StatusOK, response)
	}

	response.Status = "unhealthy"
	response.Message = fmt.Sprintf("Failed checks: %v", failures)
	return ctx.JSON(http.StatusServiceUnavailable, response)
}

// checkCookieHealth verifies cookie file can be read and written to
func (service *APIService) checkCookieHealth() string {
	cookieConfig := service.coreService.GetCookieConfig()
	if cookieConfig == nil || !cookieConfig.Enabled {
		return "disabled"
	}

	if cookieConfig.CookiePath == "" {
		return "unhealthy"
	}

	// Try to open/create file for writing (this tests write permission)
	file, err := os.OpenFile(cookieConfig.CookiePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("cookie health: cannot write to %s: %v", cookieConfig.CookiePath, err)
		return "unhealthy"
	}
	file.Close()

	// Try to read the file (this tests read permission)
	if _, err := os.ReadFile(cookieConfig.CookiePath); err != nil {
		log.Printf("cookie health: cannot read %s: %v", cookieConfig.CookiePath, err)
		return "unhealthy"
	}

	return "healthy"
}

// checkDatabaseHealth verifies database connectivity
func (service *APIService) checkDatabaseHealth() string {
	databaseService := service.coreService.GetDatabaseService()
	if databaseService == nil {
		return "unhealthy"
	}

	if _, err := databaseService.GetAllPodcastItems(); err != nil {
		log.Printf("database health: %v", err)
		return "unhealthy"
	}

	return "healthy"
}

// checkMediaHealth verifies media directory accessibility
func (service *APIService) checkMediaHealth() string {
	mediaPath := service.coreService.GetAudioSourceDirectory()
	if mediaPath == "" {
		return "unhealthy"
	}

	if err := os.MkdirAll(mediaPath, 0755); err != nil {
		log.Printf("media health: cannot create directory %s: %v", mediaPath, err)
		return "unhealthy"
	}

	return "healthy"
}
