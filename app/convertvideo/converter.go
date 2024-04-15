package convertvideo

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func ConvertVideoToAudio(videoFilePath string, outputFilePath string) error {
	return ffmpeg.Input(videoFilePath).Output(outputFilePath).Run()
}
