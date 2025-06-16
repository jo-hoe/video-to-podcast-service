package downloader

type AudioDownloader interface {
	Download(url string, path string) ([]string, error)
	IsVideoSupported(url string) bool
	IsVideoAvailable(url string) bool
}

const (
	ThumbnailUrlTag       = "WXXX" // see https://www.exiftool.org/TagNames/ID3.html for details
	PodcastDescriptionTag = "TDES"
	DateTag               = "TDA"
)
