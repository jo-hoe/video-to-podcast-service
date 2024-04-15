package download

type AudioDownloader interface {
	Download(url string, path string) ([]string, error)
	IsSupported(url string) bool
}

const ErrIsSupported = "this downloader is not responsible for this URL"
const ThumbnailUrlTag = "WXXX" // see https://www.exiftool.org/TagNames/ID3.html for details