package download

type VideoAudioDownloader interface {
	Download(url string, path string) ([]string, error)
}
