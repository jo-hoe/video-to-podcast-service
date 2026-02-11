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

func TestRootRedirectHandler_NoError(t *testing.T) {
	e := echo.New()
	mockDB := database.NewMockDatabase()
	coreService := core.NewCoreService(mockDB, "/tmp/test", nil, nil)
	uiService := NewUIService(coreService)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := uiService.rootRedirectHandler(c)
	assert.NoError(t, err)
}

func TestRootRedirectHandler_StatusMovedPermanently(t *testing.T) {
	e := echo.New()
	mockDB := database.NewMockDatabase()
	coreService := core.NewCoreService(mockDB, "/tmp/test", nil, nil)
	uiService := NewUIService(coreService)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = uiService.rootRedirectHandler(c)
	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
}

func TestRootRedirectHandler_LocationHeaderIndex(t *testing.T) {
	e := echo.New()
	mockDB := database.NewMockDatabase()
	coreService := core.NewCoreService(mockDB, "/tmp/test", nil, nil)
	uiService := NewUIService(coreService)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	_ = uiService.rootRedirectHandler(c)
	assert.Equal(t, "/index.html", rec.Header().Get("Location"))
}

func TestRootRedirectIntegration_StatusMovedPermanently(t *testing.T) {
	e := echo.New()
	mockDB := database.NewMockDatabase()
	coreService := core.NewCoreService(mockDB, "/tmp/test", nil, nil)

	uiService := NewUIService(coreService)
	uiService.SetUIRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusMovedPermanently, rec.Code)
}

func TestRootRedirectIntegration_LocationHeaderIndex(t *testing.T) {
	e := echo.New()
	mockDB := database.NewMockDatabase()
	coreService := core.NewCoreService(mockDB, "/tmp/test", nil, nil)

	uiService := NewUIService(coreService)
	uiService.SetUIRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	assert.Equal(t, "/index.html", rec.Header().Get("Location"))
}