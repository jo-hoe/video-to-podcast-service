package download

import (
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/twitch"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/youtube"
)

func TestGetVideoDownloader_ReturnsYouTubeDownloader(t *testing.T) {
	url := "https://www.youtube.com/watch?v=dQw4w9WgXcQ"

	downloader, err := GetVideoDownloader(url, nil, nil)
	if err != nil {
		t.Fatalf("GetVideoDownloader() unexpected error: %v", err)
	}
	if _, ok := downloader.(*youtube.YoutubeAudioDownloader); !ok {
		t.Fatalf("GetVideoDownloader() expected *youtube.YoutubeAudioDownloader, got %T", downloader)
	}
}

func TestGetVideoDownloader_ReturnsTwitchDownloader(t *testing.T) {
	url := "https://www.twitch.tv/videos/2345678901"

	downloader, err := GetVideoDownloader(url, nil, nil)
	if err != nil {
		t.Fatalf("GetVideoDownloader() unexpected error: %v", err)
	}
	if _, ok := downloader.(*twitch.TwitchAudioDownloader); !ok {
		t.Fatalf("GetVideoDownloader() expected *twitch.TwitchAudioDownloader, got %T", downloader)
	}
}

func TestGetVideoDownloader_UnsupportedURL_ReturnsError(t *testing.T) {
	url := "https://unsupport.com/123456789"

	downloader, err := GetVideoDownloader(url, nil, nil)
	if err == nil {
		t.Fatalf("GetVideoDownloader() expected error for unsupported url, got nil")
	}
	if downloader != nil {
		t.Fatalf("GetVideoDownloader() expected nil downloader for unsupported url, got %T", downloader)
	}
}
