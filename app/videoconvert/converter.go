package videoconvert

import (
	"errors"
	"github.com/jo-hoe/go-audio-rss-feeder/app/download"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func convertVideoToAudio(inputFilePath string, outputFilePath string) error {
	return ffmpeg.Input(inputFilePath).Output(outputFilePath).Run()
}

func VideoToAudioAssert(urlString string, outputPath string) (err error) {
	youtubeDownloader := download.YoutubeAudioDownloader{}
	if !youtubeDownloader.IsSupported(urlString) {
		return errors.New(download.ErrIsSupported)
	}

	_, err = youtubeDownloader.Download(urlString, outputPath)
	return err
}
