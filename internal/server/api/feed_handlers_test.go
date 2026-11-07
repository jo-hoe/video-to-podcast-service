package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
)

// setupFeedService builds an APIService backed by a real CoreService with one
// podcast item whose audio file actually exists on disk, so feed generation
// (which stats the file) succeeds.
func setupFeedService(t *testing.T) (*APIService, string) {
	t.Helper()
	root := t.TempDir()
	feedDir := filepath.Join(root, "channel")
	if err := os.MkdirAll(feedDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	audioPath := filepath.Join(feedDir, "episode.mp3")
	if err := os.WriteFile(audioPath, []byte("audio"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	db := database.NewMockDatabase()
	db.Items["item-1"] = &database.PodcastItem{
		ID:            "item-1",
		Title:         "Episode 1",
		AudioFilePath: audioPath,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	cs := core.NewCoreService(db, root, nil, nil, nil)
	return NewAPIService(cs, "8080"), audioPath
}

func TestFeedsHandler_ReturnsFeedList(t *testing.T) {
	e := echo.New()
	svc, _ := setupFeedService(t)

	req := httptest.NewRequest(http.MethodGet, "/"+FeedsPath, nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := svc.feedsHandler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestFeedHandler_UnknownFeed_Returns404(t *testing.T) {
	e := echo.New()
	svc, _ := setupFeedService(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle")
	ctx.SetParamValues("does-not-exist")

	err := svc.feedHandler(ctx)
	assertHTTPError(t, err, http.StatusNotFound)
}

func TestFeedHandler_MissingParam_Returns400(t *testing.T) {
	e := echo.New()
	svc, _ := setupFeedService(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	err := svc.feedHandler(ctx)
	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestAudioFileHandler_PathTraversal_Returns400(t *testing.T) {
	e := echo.New()
	svc, _ := setupFeedService(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle", "audioFileName")
	ctx.SetParamValues("..", "..")

	err := svc.audioFileHandler(ctx)
	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestAudioFileHandler_UnknownFile_Returns404(t *testing.T) {
	e := echo.New()
	svc, _ := setupFeedService(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle", "audioFileName")
	ctx.SetParamValues("channel", "not-in-db.mp3")

	err := svc.audioFileHandler(ctx)
	assertHTTPError(t, err, http.StatusNotFound)
}

func TestAudioFileHandler_ExistingFile_ServesFile(t *testing.T) {
	e := echo.New()
	svc, _ := setupFeedService(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle", "audioFileName")
	ctx.SetParamValues("channel", "episode.mp3")

	if err := svc.audioFileHandler(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
