package video

import (
	"fmt"

	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func GetMetadata(inputFilePath string) (error) {
	probeData, _ := ffmpeg.Probe(inputFilePath)
    fmt.Printf("%+v\n", (probeData))
	return nil
}
