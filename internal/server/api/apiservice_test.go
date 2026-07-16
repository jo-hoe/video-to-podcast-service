package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/labstack/echo/v4"
)

// --- equalPath ---

func TestEqualPath_SamePaths_ReturnsTrue(t *testing.T) {
	if !equalPath("/foo/bar/baz.mp3", "/foo/bar/baz.mp3") {
		t.Error("expected equal paths to match")
	}
}

func TestEqualPath_DifferentPaths_ReturnsFalse(t *testing.T) {
	if equalPath("/foo/bar/a.mp3", "/foo/bar/b.mp3") {
		t.Error("expected different paths not to match")
	}
}

func TestEqualPath_TrailingSlashNormalized(t *testing.T) {
	if !equalPath("/foo/bar/", "/foo/bar") {
		t.Error("expected trailing slash to be normalized")
	}
}

func TestEqualPath_DoubleSeparatorNormalized(t *testing.T) {
	if !equalPath("/foo//bar/baz.mp3", "/foo/bar/baz.mp3") {
		t.Error("expected double separator to be normalized")
	}
}

func TestEqualPath_DotSegmentNormalized(t *testing.T) {
	if !equalPath("/foo/./bar/baz.mp3", "/foo/bar/baz.mp3") {
		t.Error("expected dot segment to be normalized")
	}
}

func TestEqualPath_EmptyStrings_ReturnsTrue(t *testing.T) {
	if !equalPath("", "") {
		t.Error("expected two empty paths to match")
	}
}

// --- getPathAttributeValue ---

func TestGetPathAttributeValue_MissingParam_ReturnsBadRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	// no route parameter set → Param returns ""

	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)
	_, err := svc.getPathAttributeValue(ctx, "feedTitle")
	if err == nil {
		t.Fatal("expected error for missing param, got nil")
	}
	he, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if he.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", he.Code)
	}
}

func TestGetPathAttributeValue_EncodedParam_IsDecoded(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle")
	ctx.SetParamValues("My%20Feed")

	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)
	got, err := svc.getPathAttributeValue(ctx, "feedTitle")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "My Feed" {
		t.Errorf("expected %q, got %q", "My Feed", got)
	}
}

func TestGetPathAttributeValue_PlainParam_ReturnedUnchanged(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	ctx.SetParamNames("feedTitle")
	ctx.SetParamValues("MyFeed")

	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)
	got, err := svc.getPathAttributeValue(ctx, "feedTitle")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "MyFeed" {
		t.Errorf("expected %q, got %q", "MyFeed", got)
	}
}

// --- probeHandler ---

func TestProbeHandler_Returns200(t *testing.T) {
	e := echo.New()
	svc := newTestAPIService(database.NewMockDatabase(), t.TempDir(), nil)

	req := httptest.NewRequest(http.MethodGet, "/"+ProbePath, nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	if err := svc.probeHandler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}
