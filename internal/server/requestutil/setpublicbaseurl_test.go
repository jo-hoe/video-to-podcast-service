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

	// A control character makes url.Parse fail.
	if err := SetPublicBaseURL("http://exa\x7fmple.com"); err == nil {
		t.Fatal("expected error for malformed URL")
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
