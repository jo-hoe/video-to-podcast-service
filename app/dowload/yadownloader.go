package download

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/kkdai/youtube/v2"
)

const playlistRegex = `https://(?:.+)?youtube.com/(?:.+)?list=([A-Za-z0-9_-]*)`
const videoRegex = `https://(?:.+)?youtube.com/(?:.+)?watch\?v=([A-Za-z0-9_-]*)`

type YoutubeAudioDownloader struct{}

func (y *YoutubeAudioDownloader) Download(urlString string, path string) ([]string, error) {
	videoIds, err := getVideoIds(urlString)
	if err != nil {
		return make([]string, 0), err
	}
	results := make([]string, 0)
	for _, videoId := range videoIds {
		result, err := downloadVideo(videoId, path)
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

func getVideoIds(urlString string) (results []string, err error) {
	results = make([]string, 0)

	playlistId, err := getPlaylistId(urlString)
	if err == nil {
		results, err = getAllVideosInPlaylist(playlistId)
	} else {
		var videoId string
		videoId, err = getVideoId(urlString)
		if err != nil {
			return make([]string, 0), err
		}
		results = append(results, videoId)
	}

	return results, err
}

func getAllVideosInPlaylist(playlistId string) ([]string, error) {
	client := youtube.Client{}
	playlist, err := client.GetPlaylist(playlistId)
	if err != nil {
		return make([]string, 0), err
	}

	videoIds := make([]string, 0)
	for _, video := range playlist.Videos {
		videoIds = append(videoIds, video.ID)
	}

	return videoIds, nil
}

func downloadVideo(videoId string, path string) (string, error) {
	client := youtube.Client{}
	video, err := client.GetVideo(videoId)
	if err != nil {
		return "", err
	}

	formats := video.Formats.WithAudioChannels()
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		return "", err
	}
	defer stream.Close()
	return createVideoFromStream(stream, video.Title, path)
}

func createVideoFromStream(stream io.ReadCloser, videoName string, path string) (string, error) {
	fileName := fmt.Sprintf("%s.mp4", videoName)
	filePath := filepath.Join(path, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, stream)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func getVideoId(urlString string) (id string, err error) {
	return getId(urlString, videoRegex, "could not find video id in url '%s'")
}

func getPlaylistId(urlString string) (id string, err error) {
	return getId(urlString, playlistRegex, "could not find playlist id in url '%s'")
}

func getId(urlString string, regex string, errorMessage string) (id string, err error) {
	if !isValidUrl(urlString) {
		return "", fmt.Errorf("url '%s' is not a valid youtube video url", urlString)
	}
	expression := regexp.MustCompile(regex)
	matches := expression.FindStringSubmatch(urlString)
	if len(matches) == 2 {
		return matches[1], nil
	} else {
		return "", fmt.Errorf(errorMessage, urlString)
	}
}

func isValidUrl(urlString string) bool {
	_, err := url.ParseRequestURI(urlString)
	return err == nil
}
