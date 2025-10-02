package api

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

const (
	HealthPath = "health"
)

// Health check status constants
const (
	HealthStatusHealthy   = "healthy"
	HealthStatusUnhealthy = "unhealthy"
	HealthStatusDisabled  = "disabled"
)

// Health check name constants
const (
	HealthCheckService       = "service"
	HealthCheckCookies       = "cookies"
	HealthCheckDatabase      = "database"
	HealthCheckMediaDirectory = "media_directory"
)

// Health check type constants for directory checks
const (
	CheckTypeCookie = "cookie"
	CheckTypeMedia  = "media"
)

type HealthResponse struct {
	Status  string            `json:"status"`
	Checks  map[string]string `json:"checks"`
	Message string            `json:"message,omitempty"`
}

type healthCheckFunc func() (status, message string)

// healthHandler provides a comprehensive health check for Kubernetes readiness/liveness probes
func (service *APIService) healthHandler(ctx echo.Context) error {
	// Define health checks with their names and functions
	healthChecks := map[string]healthCheckFunc{
		HealthCheckService:       func() (string, string) { return HealthStatusHealthy, "" },
		HealthCheckCookies:       service.checkCookieFileHealth,
		HealthCheckDatabase:      service.checkDatabaseHealth,
		HealthCheckMediaDirectory: service.checkMediaDirectoryHealth,
	}

	checks := make(map[string]string)
	allHealthy := true
	var messages []string

	// Execute all health checks
	for checkName, checkFunc := range healthChecks {
		status, message := checkFunc()
		checks[checkName] = status
		
		if status != HealthStatusHealthy {
			allHealthy = false
			if message != "" {
				messages = append(messages, fmt.Sprintf("%s: %s", checkName, message))
			}
		}
	}

	// Build response
	response := HealthResponse{
		Checks: checks,
	}

	if allHealthy {
		response.Status = HealthStatusHealthy
		return ctx.JSON(http.StatusOK, response)
	} else {
		response.Status = HealthStatusUnhealthy
		if len(messages) > 0 {
			response.Message = fmt.Sprintf("Health check failures: %v", messages)
		}
		return ctx.JSON(http.StatusServiceUnavailable, response)
	}
}

// checkDirectoryAccess verifies directory exists and is accessible without creating test files
func (service *APIService) checkDirectoryAccess(dirPath, checkType string) (status, message string) {
	// Ensure the directory exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		log.Printf("failed to create %s directory %s: %v", checkType, dirPath, err)
		return HealthStatusUnhealthy, fmt.Sprintf("cannot create %s directory: %v", checkType, err)
	}

	// Check if directory is accessible by getting directory info
	info, err := os.Stat(dirPath)
	if err != nil {
		log.Printf("failed to access %s directory %s: %v", checkType, dirPath, err)
		return HealthStatusUnhealthy, fmt.Sprintf("cannot access %s directory: %v", checkType, err)
	}

	// Verify it's actually a directory
	if !info.IsDir() {
		log.Printf("%s path %s is not a directory", checkType, dirPath)
		return HealthStatusUnhealthy, fmt.Sprintf("%s path is not a directory", checkType)
	}

	return HealthStatusHealthy, ""
}

// checkDirectoryWriteAccess verifies directory exists, is accessible, and is writable without creating test files
func (service *APIService) checkDirectoryWriteAccess(dirPath, checkType string) (status, message string) {
	// First perform basic directory access check
	if status, message := service.checkDirectoryAccess(dirPath, checkType); status != HealthStatusHealthy {
		return status, message
	}

	// Check directory write permissions using os.Access (Unix-like systems)
	// On Windows, we'll use a different approach since os.Access isn't available
	info, err := os.Stat(dirPath)
	if err != nil {
		log.Printf("failed to stat %s directory %s: %v", checkType, dirPath, err)
		return HealthStatusUnhealthy, fmt.Sprintf("cannot stat %s directory: %v", checkType, err)
	}

	// Check if directory has write permissions
	// On Unix-like systems, check owner write permission (bit 7)
	// This is a basic check - in a real-world scenario, you might want more sophisticated permission checking
	mode := info.Mode()
	if mode.Perm()&0200 == 0 {
		log.Printf("%s directory %s is not writable", checkType, dirPath)
		return HealthStatusUnhealthy, fmt.Sprintf("%s directory is not writable", checkType)
	}

	return HealthStatusHealthy, ""
}

// checkCookieFileHealth verifies cookie file read/write accessibility without actually writing to it
func (service *APIService) checkCookieFileHealth() (status, message string) {
	cookieConfig := service.coreService.GetCookieConfig()
	if cookieConfig == nil || !cookieConfig.Enabled {
		return HealthStatusDisabled, "cookie functionality is disabled"
	}

	if cookieConfig.CookiePath == "" {
		return HealthStatusUnhealthy, "cookie path not configured"
	}

	// Check if the cookie file exists and its permissions
	if info, err := os.Stat(cookieConfig.CookiePath); err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist - check if parent directory exists and is accessible
			cookieDir := filepath.Dir(cookieConfig.CookiePath)
			if _, dirErr := os.Stat(cookieDir); dirErr != nil {
				if os.IsNotExist(dirErr) {
					log.Printf("cookie directory %s does not exist", cookieDir)
					return HealthStatusUnhealthy, "cookie directory does not exist"
				}
				log.Printf("failed to access cookie directory %s: %v", cookieDir, dirErr)
				return HealthStatusUnhealthy, fmt.Sprintf("cannot access cookie directory: %v", dirErr)
			}
			// Directory exists, file will be created when needed - this is healthy
			return HealthStatusHealthy, ""
		} else {
			// Some other error accessing the file
			log.Printf("failed to check cookie file %s: %v", cookieConfig.CookiePath, err)
			return HealthStatusUnhealthy, fmt.Sprintf("cannot access cookie file: %v", err)
		}
	} else {
		// File exists - check if it's readable and writable
		mode := info.Mode()
		
		// Check read permission (owner read bit)
		if mode.Perm()&0400 == 0 {
			log.Printf("cookie file %s is not readable", cookieConfig.CookiePath)
			return HealthStatusUnhealthy, "cookie file is not readable"
		}
		
		// Check write permission (owner write bit)
		if mode.Perm()&0200 == 0 {
			log.Printf("cookie file %s is not writable", cookieConfig.CookiePath)
			return HealthStatusUnhealthy, "cookie file is not writable"
		}
		
		// Attempt to open the file for reading to verify actual access
		file, err := os.OpenFile(cookieConfig.CookiePath, os.O_RDONLY, 0)
		if err != nil {
			log.Printf("failed to open cookie file %s for reading: %v", cookieConfig.CookiePath, err)
			return HealthStatusUnhealthy, fmt.Sprintf("cannot read cookie file: %v", err)
		}
		file.Close()
	}

	return HealthStatusHealthy, ""
}

// checkDatabaseHealth verifies database connectivity
func (service *APIService) checkDatabaseHealth() (status, message string) {
	databaseService := service.coreService.GetDatabaseService()
	if databaseService == nil {
		return HealthStatusUnhealthy, "database service not initialized"
	}

	if _, err := databaseService.GetAllPodcastItems(); err != nil {
		log.Printf("database health check failed: %v", err)
		return HealthStatusUnhealthy, fmt.Sprintf("database connectivity error: %v", err)
	}

	return HealthStatusHealthy, ""
}

// checkMediaDirectoryHealth verifies media directory accessibility
func (service *APIService) checkMediaDirectoryHealth() (status, message string) {
	mediaPath := service.coreService.GetAudioSourceDirectory()
	if mediaPath == "" {
		return HealthStatusUnhealthy, "media path not configured"
	}

	return service.checkDirectoryAccess(mediaPath, CheckTypeMedia)
}
