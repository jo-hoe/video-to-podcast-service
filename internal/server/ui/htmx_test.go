package ui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func newTestUIService(t *testing.T) *UIService {
	t.Helper()
	mockDB := database.NewMockDatabase()
	coreService := core.NewCoreService(mockDB, t.TempDir(), nil, nil, nil)
	return NewUIService(coreService)
}

func TestHtmxAddItemHandler_MissingURL_Returns400(t *testing.T) {
	e := echo.New()
	svc := newTestUIService(t)

	req := httptest.NewRequest(http.MethodPost, "/htmx/addItem", strings.NewReader("url="))
	req.Header.Set(echo.HeaderContentType, "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	_ = svc.htmxAddItemHandler(ctx)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHtmxAddItemHandler_UnsupportedURL_Returns422(t *testing.T) {
	e := echo.New()
	svc := newTestUIService(t)

	req := httptest.NewRequest(http.MethodPost, "/htmx/addItem", strings.NewReader("url=https://unsupported.example.com/video"))
	req.Header.Set(echo.HeaderContentType, "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	_ = svc.htmxAddItemHandler(ctx)
	assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
}

func TestHtmxItemsHandler_EmptyDB_Returns200(t *testing.T) {
	e := echo.New()
	mockDB := database.NewMockDatabase()
	coreService := core.NewCoreService(mockDB, t.TempDir(), nil, nil, nil)
	uiService := NewUIService(coreService)
	uiService.SetUIRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/htmx/items", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestIconHandler_Returns200(t *testing.T) {
	e := echo.New()
	svc := newTestUIService(t)
	uiService := svc
	uiService.SetUIRoutes(e)

	req := httptest.NewRequest(http.MethodGet, "/icon.svg", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Header().Get(echo.HeaderContentType), "image/svg+xml")
}
