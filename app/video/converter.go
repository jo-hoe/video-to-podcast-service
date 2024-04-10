package video

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func ConvertVideoToAudio(inputFilePath string, outputFilePath string) error {
	return ffmpeg.Input(inputFilePath).Output(outputFilePath).Run()
}
