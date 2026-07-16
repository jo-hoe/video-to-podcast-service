package downloader

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
)

func TestIsLiveFromOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   bool
	}{
		{
			name:   "single live status line",
			output: []byte("is_live"),
			want:   true,
		},
		{
			name:   "live status among multiple lines",
			output: []byte("none\nis_live\nnone"),
			want:   true,
		},
		{
			name:   "live status with surrounding whitespace",
			output: []byte("  is_live  "),
			want:   true,
		},
		{
			name:   "not live",
			output: []byte("was_live"),
			want:   false,
		},
		{
			name:   "empty output",
			output: []byte(""),
			want:   false,
		},
		{
			name:   "output with only newlines",
			output: []byte("\n\n\n"),
			want:   false,
		},
		{
			name:   "partial match is not live",
			output: []byte("is_live_stream"),
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsLiveFromOutput(tt.output); got != tt.want {
				t.Errorf("IsLiveFromOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFirstHTTPSLineFromOutput(t *testing.T) {
	tests := []struct {
		name   string
		output []byte
		want   string
	}{
		{
			name:   "single https line",
			output: []byte("https://example.com/audio.mp3"),
			want:   "https://example.com/audio.mp3",
		},
		{
			name:   "https line among multiple lines",
			output: []byte("some metadata\nhttps://example.com/audio.mp3\nmore data"),
			want:   "https://example.com/audio.mp3",
		},
		{
			name:   "returns first https line when multiple present",
			output: []byte("https://first.com\nhttps://second.com"),
			want:   "https://first.com",
		},
		{
			name:   "no https line returns empty string",
			output: []byte("http://example.com\nsome other line"),
			want:   "",
		},
		{
			name:   "empty output returns empty string",
			output: []byte(""),
			want:   "",
		},
		{
			name:   "http prefix does not match",
			output: []byte("http://example.com"),
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstHTTPSLineFromOutput(tt.output); got != tt.want {
				t.Errorf("FirstHTTPSLineFromOutput() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppendCookieArgs(t *testing.T) {
	t.Run("nil config returns args unchanged", func(t *testing.T) {
		args := []string{"--format", "mp3"}
		got := AppendCookieArgs(args, nil)
		if len(got) != len(args) {
			t.Errorf("expected %d args, got %d", len(args), len(got))
		}
	})

	t.Run("disabled cookies returns args unchanged", func(t *testing.T) {
		args := []string{"--format", "mp3"}
		cfg := &config.Cookies{Enabled: false, CookiePath: "/some/path"}
		got := AppendCookieArgs(args, cfg)
		if len(got) != len(args) {
			t.Errorf("expected %d args, got %d", len(args), len(got))
		}
	})

	t.Run("empty cookie path returns args unchanged", func(t *testing.T) {
		args := []string{"--format", "mp3"}
		cfg := &config.Cookies{Enabled: true, CookiePath: ""}
		got := AppendCookieArgs(args, cfg)
		if len(got) != len(args) {
			t.Errorf("expected %d args, got %d", len(args), len(got))
		}
	})

	t.Run("non-existent cookie file returns args unchanged", func(t *testing.T) {
		args := []string{"--format", "mp3"}
		cfg := &config.Cookies{Enabled: true, CookiePath: "/non/existent/cookies.txt"}
		got := AppendCookieArgs(args, cfg)
		if len(got) != len(args) {
			t.Errorf("expected %d args, got %d", len(args), len(got))
		}
	})

	t.Run("existing cookie file appends cookie args", func(t *testing.T) {
		cookieFile := filepath.Join(t.TempDir(), "cookies.txt")
		if err := os.WriteFile(cookieFile, []byte("# cookie data"), 0644); err != nil {
			t.Fatalf("failed to create temp cookie file: %v", err)
		}

		args := []string{"--format", "mp3"}
		cfg := &config.Cookies{Enabled: true, CookiePath: cookieFile}
		got := AppendCookieArgs(args, cfg)

		if len(got) != len(args)+2 {
			t.Fatalf("expected %d args, got %d", len(args)+2, len(got))
		}
		if got[len(got)-2] != "--cookies" {
			t.Errorf("expected --cookies flag, got %q", got[len(got)-2])
		}
		if got[len(got)-1] != cookieFile {
			t.Errorf("expected cookie path %q, got %q", cookieFile, got[len(got)-1])
		}
	})
}
