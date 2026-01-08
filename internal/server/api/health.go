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
	ProbePath  = "probe"
)

// Health status constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusDisabled  = "disabled"
)

type HealthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks"`
	Message string            `json:"message,omitempty"`
}

// healthHandler provides a health check for Kubernetes readiness/liveness probes
func (service *APIService) healthHandler(ctx echo.Context) error {
	checks := map[string]string{
		"service":         HealthStatusHealthy,
		"cookies":         service.checkCookieHealth(),
		"database":        service.checkDatabaseHealth(),
		"media_directory": service.checkMediaHealth(),
	}

	// Check if any health check failed
	allHealthy := true
	var failures []string
	for name, status := range checks {
		if status != HealthStatusHealthy && status != HealthStatusDisabled {
			allHealthy = false
			failures = append(failures, name)
		}
	}

	response := HealthResponse{Checks: checks}
	if allHealthy {
		response.Status = HealthStatusHealthy
		return ctx.JSON(http.StatusOK, response)
	}

	response.Status = HealthStatusUnhealthy
	response.Message = fmt.Sprintf("Failed checks: %v", failures)
	return ctx.JSON(http.StatusServiceUnavailable, response)
}

// checkCookieHealth verifies cookie file can be read and written to
func (service *APIService) checkCookieHealth() string {
	cookieConfig := service.coreService.GetCookieConfig()
	if cookieConfig == nil || !cookieConfig.Enabled {
		return HealthStatusDisabled
	}

	if cookieConfig.CookiePath == "" {
		return HealthStatusUnhealthy
	}

	// Try to open/create file for writing (this tests write permission)
	file, err := os.OpenFile(cookieConfig.CookiePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("cookie health: cannot write to %s: %v", cookieConfig.CookiePath, err)
		return HealthStatusUnhealthy
	}
	if err := file.Close(); err != nil {
		log.Printf("cookie health: cannot close file %s: %v", cookieConfig.CookiePath, err)
		return HealthStatusUnhealthy
	}

	// Try to read the file (this tests read permission)
	if _, err := os.ReadFile(cookieConfig.CookiePath); err != nil {
		log.Printf("cookie health: cannot read %s: %v", cookieConfig.CookiePath, err)
		return HealthStatusUnhealthy
	}

	return HealthStatusHealthy
}

// checkDatabaseHealth verifies database connectivity
func (service *APIService) checkDatabaseHealth() string {
	databaseService := service.coreService.GetDatabaseService()
	if databaseService == nil {
		return HealthStatusUnhealthy
	}

	if _, err := databaseService.GetAllPodcastItems(); err != nil {
		log.Printf("database health: %v", err)
		return HealthStatusUnhealthy
	}

	return HealthStatusHealthy
}

// checkMediaHealth verifies media directory accessibility
func (service *APIService) checkMediaHealth() string {
	mediaPath := service.coreService.GetAudioSourceDirectory()
	if mediaPath == "" {
		return HealthStatusUnhealthy
	}

	if err := os.MkdirAll(mediaPath, 0755); err != nil {
		log.Printf("media health: cannot create directory %s: %v", mediaPath, err)
		return HealthStatusUnhealthy
	}

	return HealthStatusHealthy
}
