package ui

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestRootRedirectHandler(t *testing.T) {
	// Setup
	e := echo.New()

	// Create a mock database service
	mockDB := database.NewMockDatabase()

	// Create core service with mock database
	coreService := core.NewCoreService(mockDB, "/tmp/test", nil, nil)

	// Create UI service
	uiService := NewUIService(coreService)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute the handler
	err := uiService.rootRedirectHandler(c)

	// Assert no error occurred
	assert.NoError(t, err)

	// Assert the response status code is 301 (Moved Permanently)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)

	// Assert the Location header points to /index.html
	assert.Equal(t, "/index.html", rec.Header().Get("Location"))
}

func TestRootRedirectIntegration(t *testing.T) {
	// Setup
	e := echo.New()

	// Create a mock database service
	mockDB := database.NewMockDatabase()

	// Create core service with mock database
	coreService := core.NewCoreService(mockDB, "/tmp/test", nil, nil)

	// Create UI service and set up routes
	uiService := NewUIService(coreService)
	uiService.SetUIRoutes(e)

	// Create test request for root path
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	// Serve the request
	e.ServeHTTP(rec, req)

	// Assert the response status code is 301
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)

	// Assert the Location header points to /index.html
	assert.Equal(t, "/index.html", rec.Header().Get("Location"))
}
