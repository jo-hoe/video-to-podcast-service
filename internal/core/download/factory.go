package download

import (
	"fmt"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/youtube"
)

const ErrIsVideoSupported = "this downloader is not responsible for this URL '%s'"

func GetVideoDownloader(url string, cookiesConfig *config.Cookies, mediaConfig *config.Media) (downloader downloader.AudioDownloader, err error) {
	youtubeAudioDownloader := youtube.NewYoutubeAudioDownloader(cookiesConfig, mediaConfig)
	if youtubeAudioDownloader.IsVideoSupported(url) {
		return youtubeAudioDownloader, nil
	}

	return nil, fmt.Errorf(ErrIsVideoSupported, url)
}
