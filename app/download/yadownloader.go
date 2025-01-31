package download

import (
	"fmt"
	"os"
	"path/filepath"
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
	dl := ytdlp.New().ExtractAudio().AudioFormat("mp3").Output(tempFilenameTemplate)
	_, err := dl.Run(context.Background(), urlString)
	if err != nil {
		return make([]string, 0), err
	}

	// get all video file paths which have been downloaded
	var downloadedFiles []string
	err = filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".mp3") {
			downloadedFiles = append(downloadedFiles, path)
		}
		return nil
	})
	if err != nil {
		return make([]string, 0), err
	}

	return downloadedFiles, nil
}

func (y *YoutubeAudioDownloader) IsVideoSupported(url string) bool {
	return regexp.MustCompile(playlistRegex).MatchString(url) || regexp.MustCompile(videoRegex).MatchString(url) || regexp.MustCompile(videoShortRegex).MatchString(url)
}

func (y *YoutubeAudioDownloader) IsVideoAvailable(urlString string) bool {
	dl, err := ytdlp.New().Simulate().Quiet().Run(context.Background(), urlString)
	if err != nil {
		return false
	}
	return dl != nil
}
