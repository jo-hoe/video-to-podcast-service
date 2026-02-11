package downloader

type AudioDownloader interface {
	// Download downloads the audio from the given URL and saves it to the specified path.
	// It returns a slice of file paths that were downloaded as the url may point to multiple files (e.g., a playlist).
	// The downloader decides if subpath are created or not.
	// Subpath are not guaranteed but used for example to group channels or playlists.
	Download(url string, path string) ([]string, error)
	IsVideoSupported(url string) bool
	IsVideoAvailable(url string) bool
	// ListVideoEntries returns individual video URLs for a given input URL.
	// For playlist URLs, it returns all video URLs in the playlist.
	// For single video URLs, it returns a slice containing the original URL.
	ListVideoEntries(url string) ([]string, error)
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
