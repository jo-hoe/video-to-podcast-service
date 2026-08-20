package downloader

import "errors"

// ErrVideoLive is returned by CheckVideoAvailability when the content is
// currently streaming live. Callers should not retry immediately.
var ErrVideoLive = errors.New("video is currently live")

type AudioDownloader interface {
	// Download downloads the audio from a single video URL and saves it to the specified path.
	// It returns the full file path to the downloaded audio file.
	// The downloader decides if subpaths are created or not.
	Download(url string, path string) (string, error)
	IsVideoSupported(url string) bool
	// CheckVideoAvailability returns nil if the video is available for download,
	// ErrVideoLive if it is currently live, or another error if unavailable.
	CheckVideoAvailability(url string) error
	// ListIndividualVideoURLs returns individual video URLs for a given input URL.
	// For playlist URLs, it returns all video URLs in the playlist.
	// For single video URLs, it returns a slice containing the original URL.
	ListIndividualVideoURLs(url string) ([]string, error)
}

const (
	ThumbnailUrlTag       = "WXXX" // see https://www.exiftool.org/TagNames/ID3.html for details
	PodcastDescriptionTag = "TDES"
	DateTag               = "TDA"
	Title                 = "title"
	Artist                = "artist"

	// custom tags
	VideoDownloadLink = "videolink"
)
