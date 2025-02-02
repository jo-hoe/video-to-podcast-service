package download

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/jo-hoe/video-to-podcast-service/app/filemanagement"
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
	results := make([]string, 0)
	// create temp directory
	tempPath, err := os.MkdirTemp("", "")
	if err != nil {
		return results, err
	}
	defer os.RemoveAll(tempPath)

	tempResults, err := download(tempPath, urlString)
	if err != nil {
		return nil, err
	}

	for _, fullFilePath := range tempResults {
		directoryName := filepath.Base(filepath.Dir(fullFilePath))
		targetSubDirectory := filepath.Join(path, directoryName)
		err = os.MkdirAll(targetSubDirectory, os.ModePerm)
		if err != nil {
			return nil, err
		}

		targetFilename := filepath.Base(fullFilePath)
		targetPath := filepath.Join(targetSubDirectory, targetFilename)
		err = filemanagement.MoveFile(fullFilePath, targetPath)
		if err != nil {
			return nil, err
		}
		results = append(results, targetPath)
	}

	return results, err
}

func download(targetDirectory string, urlString string) ([]string, error) {
	result := make([]string, 0)
	// set download behavior
	tempFilenameTemplate := fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, "%(channel)s/%(title)s.%(ext)s")
	dl := ytdlp.New().
		ExtractAudio().AudioFormat("mp3"). // convert get mp3 after downloading the video
		EmbedMetadata().                   // adds metadata such as artist to the file
		Output(tempFilenameTemplate)       // set output path

	// download
	log.Printf("downloading from '%s' to '%s'", urlString, targetDirectory)
	_, err := dl.Run(context.Background(), urlString)
	if err != nil {
		return result, err
	}
	log.Printf("completed downloaded from '%s' to '%s'", urlString, targetDirectory)

	// get file names
	result, err = filemanagement.GetAudioFiles(targetDirectory)
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
