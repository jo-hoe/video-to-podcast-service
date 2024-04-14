package videoconvert

import (
	"errors"

	"github.com/jo-hoe/go-audio-rss-feeder/app/download"
	"github.com/jo-hoe/go-audio-rss-feeder/app/video"
	mp3joiner "github.com/jo-hoe/mp3-joiner"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func convertVideoToAudio(videoFilePath string, outputFilePath string) error {
	// preserve metadata
	metadata, err := video.GetTagMetadata(videoFilePath)
	if err != nil {
		return err
	}

	// actual convert set
	if err := ffmpeg.Input(videoFilePath).Output(outputFilePath).Run(); err != nil {
		return err
	}

	// re-apply metadata
	return mp3joiner.SetFFmpegMetadataTag(outputFilePath, metadata, make([]mp3joiner.Chapter, 0))
}

func ConvertVideoToAudio(urlString string, outputPath string) (err error) {
	youtubeDownloader := download.YoutubeAudioDownloader{}
	if !youtubeDownloader.IsSupported(urlString) {
		return errors.New(download.ErrIsSupported)
	}

	videos, err := youtubeDownloader.Download(urlString, outputPath)
	for _, videoItem := range videos {
		if err := convertVideoToAudio(videoItem, outputPath); err != nil {
			return err
		}
	}
	return err
}
