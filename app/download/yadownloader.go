package download

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jo-hoe/go-audio-rss-feeder/app/convertvideo"
	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/kkdai/youtube/v2"
)

const playlistRegex = `https://(?:.+)?youtube.com/(?:.+)?list=([A-Za-z0-9_-]*)`
const videoRegex = `https://(?:.+)?youtube.com/(?:.+)?watch\?v=([A-Za-z0-9_-]*)`

type YoutubeAudioDownloader struct{}

func (y *YoutubeAudioDownloader) Download(urlString string, path string) ([]string, error) {
	videosMetadata, err := getAllYoutubeMetadata(urlString)
	if err != nil {
		return make([]string, 0), err
	}
	results := make([]string, 0)
	for _, videoMetadata := range videosMetadata {
		videoFile, err := downloadVideo(videoMetadata, os.TempDir())
		defer os.Remove(videoFile)
		if err != nil {
			return results, err
		}

		// create author specific path if not exist
		calculatedPath := filepath.Join(path, videoMetadata.Author)
		err = os.MkdirAll(calculatedPath, os.ModePerm)
		if err != nil {
			return results, err
		}

		audioPath, err := convertToAudio(videoFile, videoMetadata, calculatedPath)
		if err != nil {
			return results, err
		}
		results = append(results, audioPath)
	}
	return results, nil
}

func convertToAudio(videoFile string, youtubeMetadata *youtube.Video, path string) (string, error) {
	fileName := filepath.Base(videoFile)
	fileNameWithoutExtension := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	audioFile := filepath.Join(path, fmt.Sprintf("%s.mp3", fileNameWithoutExtension))

	err := convertvideo.ConvertVideoToAudio(videoFile, audioFile)
	if err != nil {
		return "", err
	}
	metadata := getAudioMetaData(youtubeMetadata)
	err = mp3joiner.SetFFmpegMetadataTag(audioFile, metadata, make([]mp3joiner.Chapter, 0))
	if err != nil {
		return "", err
	}
	return audioFile, nil
}

func getAudioMetaData(youtubeMetadata *youtube.Video) map[string]string {
	result := make(map[string]string)

	thumbnailMaxSize := uint(0)
	for i, thumbnail := range youtubeMetadata.Thumbnails {
		if thumbnail.Width > thumbnailMaxSize {
			result[ThumbnailUrlTag] = youtubeMetadata.Thumbnails[i].URL
		}

	}
	result["Artist"] = youtubeMetadata.Author
	result["Title"] = youtubeMetadata.Title
	result["Comment"] = youtubeMetadata.Description

	return result
}

func (y *YoutubeAudioDownloader) IsSupported(url string) bool {
	return regexp.MustCompile(playlistRegex).MatchString(url) || regexp.MustCompile(videoRegex).MatchString(url)
}

func getAllYoutubeMetadata(urlString string) (results []*youtube.Video, err error) {
	results = make([]*youtube.Video, 0)

	playlistId, err := getPlaylistId(urlString)
	if err == nil {
		results, err = getAllYoutubeMetadataInPlaylist(playlistId)
	} else {
		var videoId string
		videoId, err = getVideoId(urlString)
		if err != nil {
			return make([]*youtube.Video, 0), err
		}
		var video *youtube.Video
		video, err = getYoutubeMetadata(videoId)
		if err != nil {
			return make([]*youtube.Video, 0), err
		}
		results = append(results, video)
	}

	return results, err
}

func getYoutubeMetadata(videoId string) (video *youtube.Video, err error) {
	client := youtube.Client{}
	video, err = client.GetVideo(videoId)
	if err != nil {
		return nil, err
	}

	return video, nil
}

func getAllYoutubeMetadataInPlaylist(playlistId string) ([]*youtube.Video, error) {
	client := youtube.Client{}
	playlist, err := client.GetPlaylist(playlistId)
	if err != nil {
		return make([]*youtube.Video, 0), err
	}

	videos := make([]*youtube.Video, 0)
	for _, video := range playlist.Videos {
		video, err := getYoutubeMetadata(video.ID)
		if err != nil {
			return make([]*youtube.Video, 0), err
		}
		videos = append(videos, video)
	}

	return videos, nil
}

func downloadVideo(video *youtube.Video, path string) (string, error) {
	client := youtube.Client{}
	formats := video.Formats.WithAudioChannels()
	// TODO: formats define the quality
	// come up with a better strategy to select the best format
	stream, _, err := client.GetStream(video, &formats[0])
	if err != nil {
		return "", err
	}
	defer stream.Close()
	path, err = createVideoFromStream(stream, video.Title, path)
	if err != nil {
		return "", err
	}

	return path, nil
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
