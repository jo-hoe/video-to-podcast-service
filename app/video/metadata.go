package video

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func GetMetadata(inputFilePath string) (string, error) {
	probeData, _ := ffmpeg.Probe(inputFilePath)
	return probeData, nil
}
