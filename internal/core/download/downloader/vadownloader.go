package downloader

type AudioDownloader interface {
	// Download downloads the audio from a single video URL and saves it to the specified path.
	// It returns the full file path to the downloaded audio file.
	// The downloader decides if subpaths are created or not.
	Download(url string, path string) (string, error)
	IsVideoSupported(url string) bool
	IsVideoAvailable(url string) bool
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
