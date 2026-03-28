package downloader

import (
	"log/slog"
	"os"
	"strings"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
)

const (
	// LiveStatusKey is the yt-dlp field name for the live status of a video.
	LiveStatusKey = "live_status"
	// LiveStatusLiveValue is the yt-dlp live_status value for an active live stream.
	LiveStatusLiveValue = "is_live"
	// VideoURLID3Key is the ID3 tag yt-dlp uses to store the original video URL.
	VideoURLID3Key = "purl"
)

// AppendCookieArgs appends --cookies <path> to args when cookiesConfig is enabled
// and the cookie file exists on disk.
func AppendCookieArgs(args []string, cookiesConfig *config.Cookies) []string {
	if cookiesConfig == nil || !cookiesConfig.Enabled || cookiesConfig.CookiePath == "" {
		return args
	}
	if _, err := os.Stat(cookiesConfig.CookiePath); err == nil {
		slog.Info("using cookie file path", "path", cookiesConfig.CookiePath)
		return append(args, "--cookies", cookiesConfig.CookiePath)
	}
	slog.Warn("cookie file path specified but not found", "path", cookiesConfig.CookiePath)
	return args
}

// IsLiveFromOutput returns true if any non-empty line in the yt-dlp --print live_status
// output equals LiveStatusLiveValue, indicating the content is currently streaming live.
func IsLiveFromOutput(output []byte) bool {
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.TrimSpace(line) == LiveStatusLiveValue {
			return true
		}
	}
	return false
}

// FirstHTTPSLineFromOutput returns the first line starting with "https" from
// yt-dlp --print output, or an empty string if none is found.
func FirstHTTPSLineFromOutput(output []byte) string {
	for _, line := range strings.Split(strings.TrimSpace(string(output)), "\n") {
		if strings.HasPrefix(line, "https") {
			return line
		}
	}
	return ""
}
