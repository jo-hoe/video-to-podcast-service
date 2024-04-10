package download

type VideoAudioDownloader interface {
	Download(url string, path string) ([]string, error)
	IsSupported(url string) bool
}

const ErrIsSupported = "this downloader is not responsible for this URL"
const ThumbnailUrlTag = "comment"