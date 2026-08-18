package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
)

func newTestAPIService(db database.DatabaseService, mediaDir string, cookies *config.Cookies) *APIService {
	mediaConfig := &config.Media{MediaPath: mediaDir}
	coreService := core.NewCoreService(db, mediaDir, cookies, mediaConfig, nil)
	return NewAPIService(coreService, "8080")
}

// --- checkCookieHealth ---

func TestCheckCookieHealth_NilConfig_ReturnsDisabled(t *testing.T) {
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)
	if got := svc.checkCookieHealth(); got != HealthStatusDisabled {
		t.Errorf("expected %q, got %q", HealthStatusDisabled, got)
	}
}

func TestCheckCookieHealth_DisabledConfig_ReturnsDisabled(t *testing.T) {
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), &config.Cookies{Enabled: false})
	if got := svc.checkCookieHealth(); got != HealthStatusDisabled {
		t.Errorf("expected %q, got %q", HealthStatusDisabled, got)
	}
}

func TestCheckCookieHealth_EmptyPath_ReturnsUnhealthy(t *testing.T) {
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), &config.Cookies{Enabled: true, CookiePath: ""})
	if got := svc.checkCookieHealth(); got != HealthStatusUnhealthy {
		t.Errorf("expected %q, got %q", HealthStatusUnhealthy, got)
	}
}

func TestCheckCookieHealth_ValidFile_ReturnsHealthy(t *testing.T) {
	cookieFile := filepath.Join(t.TempDir(), "cookies.txt")
	if err := os.WriteFile(cookieFile, []byte(""), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), &config.Cookies{Enabled: true, CookiePath: cookieFile})
	if got := svc.checkCookieHealth(); got != HealthStatusHealthy {
		t.Errorf("expected %q, got %q", HealthStatusHealthy, got)
	}
}

func TestCheckCookieHealth_NonExistentParentDir_ReturnsUnhealthy(t *testing.T) {
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), &config.Cookies{
		Enabled:    true,
		CookiePath: "/non/existent/path/cookies.txt",
	})
	if got := svc.checkCookieHealth(); got != HealthStatusUnhealthy {
		t.Errorf("expected %q, got %q", HealthStatusUnhealthy, got)
	}
}

// --- checkDatabaseHealth ---

func TestCheckDatabaseHealth_WorkingDatabase_ReturnsHealthy(t *testing.T) {
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)
	if got := svc.checkDatabaseHealth(); got != HealthStatusHealthy {
		t.Errorf("expected %q, got %q", HealthStatusHealthy, got)
	}
}

func TestCheckDatabaseHealth_FailingDatabase_ReturnsUnhealthy(t *testing.T) {
	mock := database.NewMockDatabase()
	mock.GetAllPodcastItemsFunc = func() ([]*database.PodcastItem, error) {
		return nil, errors.New("db connection error")
	}
	svc := newTestAPIService(mock, t.TempDir(), nil)
	if got := svc.checkDatabaseHealth(); got != HealthStatusUnhealthy {
		t.Errorf("expected %q, got %q", HealthStatusUnhealthy, got)
	}
}

// --- checkMediaHealth ---

func TestCheckMediaHealth_ExistingDirectory_ReturnsHealthy(t *testing.T) {
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)
	if got := svc.checkMediaHealth(); got != HealthStatusHealthy {
		t.Errorf("expected %q, got %q", HealthStatusHealthy, got)
	}
}

func TestCheckMediaHealth_EmptyPath_ReturnsUnhealthy(t *testing.T) {
	svc := newTestAPIService(database.NewMockDatabase(), "", nil)
	if got := svc.checkMediaHealth(); got != HealthStatusUnhealthy {
		t.Errorf("expected %q, got %q", HealthStatusUnhealthy, got)
	}
}

func TestCheckMediaHealth_NewDirectory_CreatesAndReturnsHealthy(t *testing.T) {
	newDir := filepath.Join(t.TempDir(), "new_media_dir")
	svc := newTestAPIService(database.NewMockDatabase(), newDir, nil)
	if got := svc.checkMediaHealth(); got != HealthStatusHealthy {
		t.Errorf("expected %q, got %q", HealthStatusHealthy, got)
	}
}

// --- healthHandler integration ---

func TestHealthHandler_AllHealthy_Returns200(t *testing.T) {
	e := echo.New()
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)

	req := httptest.NewRequest(http.MethodGet, "/"+HealthPath, nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := svc.healthHandler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Status != HealthStatusHealthy {
		t.Errorf("expected status %q, got %q", HealthStatusHealthy, resp.Status)
	}
}

func TestHealthHandler_FailingDatabase_Returns503(t *testing.T) {
	e := echo.New()
	mock := database.NewMockDatabase()
	mock.GetAllPodcastItemsFunc = func() ([]*database.PodcastItem, error) {
		return nil, errors.New("db error")
	}
	svc := newTestAPIService(mock, t.TempDir(), nil)

	req := httptest.NewRequest(http.MethodGet, "/"+HealthPath, nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := svc.healthHandler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rec.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Status != HealthStatusUnhealthy {
		t.Errorf("expected status %q, got %q", HealthStatusUnhealthy, resp.Status)
	}
}
