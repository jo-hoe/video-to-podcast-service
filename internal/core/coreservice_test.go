package core

import (
	"net/url"
	"path/filepath"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
)

func mustParseURL(t *testing.T, raw string) *url.URL {
	t.Helper()
	u, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("failed to parse URL %q: %v", raw, err)
	}
	return u
}

func newTestCoreService(audioDir string) *CoreService {
	return NewCoreService(database.NewMockDatabase(), audioDir, nil, nil, nil)
}

func TestGetFeedDirectory_EmptyPath_ReturnsError(t *testing.T) {
	cs := newTestCoreService(filepath.FromSlash("/media"))
	_, err := cs.GetFeedDirectory("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestGetFeedDirectory_TooShallowPath_ReturnsError(t *testing.T) {
	root := filepath.FromSlash("/media")
	cs := newTestCoreService(root)
	_, err := cs.GetFeedDirectory(filepath.Join(root, "file.mp3"))
	if err == nil {
		t.Fatal("expected error for path with no parent folder")
	}
}

func TestGetFeedDirectory_ValidPath_ReturnsParentDir(t *testing.T) {
	root := filepath.FromSlash("/media")
	cs := newTestCoreService(root)
	got, err := cs.GetFeedDirectory(filepath.Join(root, "channel", "file.mp3"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "channel" {
		t.Errorf("expected %q, got %q", "channel", got)
	}
}

func TestGetLinkToFeed_ValidPath_ReturnsURL(t *testing.T) {
	root := filepath.FromSlash("/media")
	cs := newTestCoreService(root)
	u := mustParseURL(t, "http://localhost:8080")
	got := cs.GetLinkToFeed(u, "v1/feeds", filepath.Join(root, "channel", "file.mp3"))
	if got == "" {
		t.Fatal("expected non-empty link")
	}
}

func TestGetLinkToAudioFile_ValidPath_ReturnsURL(t *testing.T) {
	root := filepath.FromSlash("/media")
	cs := newTestCoreService(root)
	u := mustParseURL(t, "http://localhost:8080")
	got := cs.GetLinkToAudioFile(u, "v1/feeds", filepath.Join(root, "channel", "file.mp3"))
	if got == "" {
		t.Fatal("expected non-empty link")
	}
}

func TestGetAudioSourceDirectory(t *testing.T) {
	dir := filepath.FromSlash("/media/audio")
	cs := newTestCoreService(dir)
	if got := cs.GetAudioSourceDirectory(); got != dir {
		t.Errorf("expected %q, got %q", dir, got)
	}
}

func TestGetDatabaseService(t *testing.T) {
	db := database.NewMockDatabase()
	cs := NewCoreService(db, filepath.FromSlash("/media"), nil, nil, nil)
	if cs.GetDatabaseService() != db {
		t.Error("expected same database service")
	}
}

func TestGetCookieConfig_NilReturnsNil(t *testing.T) {
	cs := newTestCoreService(filepath.FromSlash("/media"))
	if cs.GetCookieConfig() != nil {
		t.Error("expected nil cookie config")
	}
}
