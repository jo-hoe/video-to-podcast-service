package download

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/lrstanley/go-ytdlp"
	"golang.org/x/net/context"
)

const playlistRegex = `https://(?:.+)?youtube.com/(?:.+)?list=([A-Za-z0-9_-]*)`
const videoRegex = `https://(?:.+)?youtube.com/(?:.+)?watch\?v=([A-Za-z0-9_-]*)`
const videoShortRegex = `https://youtu\.be/([A-Za-z0-9_-]*)`

type YoutubeAudioDownloader struct{}

func NewYoutubeAudioDownloader() *YoutubeAudioDownloader {
	ytdlp.MustInstall(context.Background(), nil)

	return &YoutubeAudioDownloader{}
}

func (y *YoutubeAudioDownloader) Download(urlString string, path string) ([]string, error) {
	tempFilenameTemplate := fmt.Sprintf("%s%c%s", path, os.PathSeparator, "%(channel)s/%(title)s.%(ext)s")
	dl := ytdlp.New().
		ExtractAudio().AudioFormat("mp3"). // convert get mp3 after downloading the video
		EmbedMetadata().                   // adds metadata such as artist to the file
		Output(tempFilenameTemplate)       // set output path

	log.Printf("downloading from '%s' to '%s'", urlString, path)
	cliOutput, err := dl.Run(context.Background(), urlString)
	if err != nil {
		log.Printf("error downloading from '%s' to '%s': '%v'", urlString, path, err)
		return make([]string, 0), err
	}
	log.Printf("completed downloaded from '%s' to '%s'", urlString, path)

	result := make([]string, 0)
	for _, output := range cliOutput.OutputLogs {
		// Expect the output to be in the format
		// "[ExtractAudio] Destination: <path>\\<channel name>\\<file name>.mp3"
		if strings.HasPrefix(output.Line, "[ExtractAudio] Destination: ") {
			result = append(result, strings.TrimPrefix(output.Line, "[ExtractAudio] Destination: "))
		}
	}

	log.Printf("downloaded files: %v", result)

	return result, err
}

func (y *YoutubeAudioDownloader) IsVideoSupported(url string) bool {
	return regexp.MustCompile(playlistRegex).MatchString(url) || regexp.MustCompile(videoRegex).MatchString(url) || regexp.MustCompile(videoShortRegex).MatchString(url)
}

func (y *YoutubeAudioDownloader) IsVideoAvailable(urlString string) bool {
	dl, err := ytdlp.New().Simulate().Quiet().Run(context.Background(), urlString)
	if err != nil {
		log.Printf("error checking video availability: '%v'", err)
		return false
	}
	return dl != nil
}
