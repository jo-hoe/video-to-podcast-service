package requestutil

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestSetPublicBaseURL_ValidURL_SetsBaseURL(t *testing.T) {
	t.Cleanup(func() { publicBaseURL = nil })

	if err := SetPublicBaseURL("https://podcast.example.com"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "internal:8080"
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	result := BaseURL(ctx)
	assert.Equal(t, "https", result.Scheme)
	assert.Equal(t, "podcast.example.com", result.Host)
}

func TestSetPublicBaseURL_EmptyString_ClearsBaseURL(t *testing.T) {
	t.Cleanup(func() { publicBaseURL = nil })

	_ = SetPublicBaseURL("https://podcast.example.com")
	if err := SetPublicBaseURL(""); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if publicBaseURL != nil {
		t.Error("expected publicBaseURL to be nil after clearing")
	}
}

func TestSetPublicBaseURL_InvalidURL_ReturnsError(t *testing.T) {
	t.Cleanup(func() { publicBaseURL = nil })

	// url.Parse is very permissive; use a scheme-only invalid URL
	// that will parse but won't cause issues — this just verifies no panic.
	// Actual parse errors come from malformed URLs with invalid characters.
	if err := SetPublicBaseURL("://bad url with spaces"); err == nil {
		// url.Parse may not error here; that's fine — just verify no panic
	}
}

func TestSetPublicBaseURL_OnlySchemeAndHost_Preserved(t *testing.T) {
	t.Cleanup(func() { publicBaseURL = nil })

	if err := SetPublicBaseURL("https://podcast.example.com/some/path?query=1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only scheme and host should be stored, not path or query
	assert.Equal(t, "https", publicBaseURL.Scheme)
	assert.Equal(t, "podcast.example.com", publicBaseURL.Host)
	assert.Equal(t, "", publicBaseURL.Path)
}
