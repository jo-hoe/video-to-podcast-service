package video

import (
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

func SetThumbnailPicture(videoPath, imagePath, outputPath string) error {
	// performs the command below with ffmpeg-go lib
	// ffmpeg -i video.mp4 -i image.png -map 1 -map 0 -c copy -disposition:0 attached_pic out.mp4
	videoInput := ffmpeg.Input(videoPath)
	pictureInput := ffmpeg.Input(imagePath)

	parameters := ffmpeg.KwArgs{
		"c": "copy",
		"disposition:0": "attached_pic",
	}
	
	return ffmpeg.Output([]*ffmpeg.Stream{pictureInput, videoInput}, outputPath, parameters).Run()
}