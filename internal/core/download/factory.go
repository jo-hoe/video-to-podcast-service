package download

import (
	"fmt"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/youtube"
)

const ErrIsVideoSupported = "this downloader is not responsible for this URL '%s'"

func GetVideoDownloader(url string) (downloader downloader.AudioDownloader, err error) {
	youtubeAudioDownloader := &youtube.YoutubeAudioDownloader{}
	if youtubeAudioDownloader.IsVideoSupported(url) {
		return youtubeAudioDownloader, nil
	}

	return nil, fmt.Errorf(ErrIsVideoSupported, url)
}
