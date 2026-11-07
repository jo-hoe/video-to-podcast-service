package api

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
)

func TestDeleteFeedItem_MissingID_Returns400(t *testing.T) {
	e := echo.New()
	svc := newTestAPIService(newMockService())

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle", "podcastItemID")
	ctx.SetParamValues("channel", "")

	err := svc.deleteFeedItem(ctx)
	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestDeleteFeedItem_UnknownID_Returns400(t *testing.T) {
	e := echo.New()
	svc := newTestAPIService(newMockService())

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle", "podcastItemID")
	ctx.SetParamValues("channel", "unknown-id")

	err := svc.deleteFeedItem(ctx)
	assertHTTPError(t, err, http.StatusBadRequest)
}

func TestDeleteFeedItem_ValidItem_Returns200(t *testing.T) {
	e := echo.New()
	mock := newMockService()

	item := &database.PodcastItem{
		ID:            "item-1",
		AudioFilePath: filepath.Join("channel", "file.mp3"),
	}
	mock.DatabaseService.(*database.MockDatabase).Items["item-1"] = item
	mock.GetFeedDirectoryFunc = func(_ string) (string, error) { return "channel", nil }
	mock.DeletePodcastItemFunc = func(_ string) error { return nil }

	svc := newTestAPIService(mock)

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle", "podcastItemID")
	ctx.SetParamValues("channel", "item-1")

	if err := svc.deleteFeedItem(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestIsWithinRoot_InsideRoot_ReturnsTrue(t *testing.T) {
	root := filepath.Clean(filepath.FromSlash("/media"))
	inside := filepath.Join(root, "channel", "file.mp3")
	if !isWithinRoot(inside, root) {
		t.Error("expected path inside root to return true")
	}
}

func TestIsWithinRoot_OutsideRoot_ReturnsFalse(t *testing.T) {
	root := filepath.Clean(filepath.FromSlash("/media"))
	outside := filepath.Join(filepath.FromSlash("/other"), "channel", "file.mp3")
	if isWithinRoot(outside, root) {
		t.Error("expected path outside root to return false")
	}
}

func TestIsWithinRoot_ExactRoot_ReturnsFalse(t *testing.T) {
	root := filepath.Clean(filepath.FromSlash("/media"))
	if isWithinRoot(root, root) {
		t.Error("expected exact root match to return false")
	}
}

func assertHTTPError(t *testing.T, err error, code int) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != code {
		t.Errorf("expected %d, got %d", code, he.Code)
	}
}
