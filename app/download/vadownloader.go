package download

import (
	"fmt"
	"regexp"
	"strings"
)

type AudioDownloader interface {
	Download(url string, path string) ([]string, error)
	IsVideoSupported(url string) bool
	IsVideoAvailable(url string) bool
}

const (
	ErrIsVideoSupported   = "this downloader is not responsible for this URL '%s'"
	ThumbnailUrlTag       = "WXXX" // see https://www.exiftool.org/TagNames/ID3.html for details
	PodcastDescriptionTag = "TDES"
)

func GetVideoDownloader(url string) (downloader AudioDownloader, err error) {
	youtubeAudioDownloader := &YoutubeAudioDownloader{}
	if youtubeAudioDownloader.IsVideoSupported(url) {
		return youtubeAudioDownloader, nil
	}

	return nil, fmt.Errorf(ErrIsVideoSupported, url)
}

func sanitizeFilename(filename string) string {
	// Define a regex pattern for invalid characters.
	// This pattern includes:
	// - ASCII control characters (0-31 and 127)
	// - Characters forbidden in Windows filenames: \ / : * ? " < > |
	// - Characters typically forbidden in Unix filenames: /
	// - Characters generally not used in filenames: non-printable characters
	invalidChars := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1F\x7F]`)

	// Replace invalid characters with underscores.
	sanitized := invalidChars.ReplaceAllString(filename, "_")

	// Additional cleanup:
	// - Remove leading/trailing whitespace
	// - Collapse consecutive underscores into a single underscore
	sanitized = strings.TrimSpace(sanitized)
	sanitized = regexp.MustCompile(`_+`).ReplaceAllString(sanitized, "_")

	return sanitized
}
