package download

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/jo-hoe/video-to-podcast-service/app/convertvideo"
	"github.com/jo-hoe/video-to-podcast-service/app/filemanagement"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/kkdai/youtube/v2"
)

const playlistRegex = `https://(?:.+)?youtube.com/(?:.+)?list=([A-Za-z0-9_-]*)`
const videoRegex = `https://(?:.+)?youtube.com/(?:.+)?watch\?v=([A-Za-z0-9_-]*)`
const videoShortRegex = `https://youtu\.be/([A-Za-z0-9_-]*)`

type YoutubeAudioDownloader struct{}

func (y *YoutubeAudioDownloader) Download(urlString string, path string) ([]string, error) {
	videosMetadata, err := getAllYoutubeMetadata(urlString)
	if err != nil {
		return make([]string, 0), err
	}
	results := make([]string, 0)
	for _, videoMetadata := range videosMetadata {
		videoFilePath, err := downloadVideo(videoMetadata, os.TempDir())
		if err != nil {
			return results, err
		}

		defer func() {
			log.Printf("deleting video file '%s'\n", videoFilePath)
			err := os.Remove(videoFilePath)
			if err != nil {
				log.Printf("error: %v when removing video file '%s'\n", err, videoFilePath)
			}
		}()

		// create author specific path if not exist
		calculatedPath := filepath.Join(path, videoMetadata.Author)
		err = os.MkdirAll(calculatedPath, os.ModePerm)
		if err != nil {
			return results, err
		}

		audioPath, err := convertToAudio(videoFilePath, videoMetadata, calculatedPath)
		if err != nil {
			return results, err
		}
		results = append(results, audioPath)
	}
	return results, nil
}

func (y *YoutubeAudioDownloader) IsVideoSupported(url string) bool {
	return regexp.MustCompile(playlistRegex).MatchString(url) || regexp.MustCompile(videoRegex).MatchString(url) || regexp.MustCompile(videoShortRegex).MatchString(url)
}

func (y *YoutubeAudioDownloader) IsVideoAvailable(urlString string) bool {
	_, err := getAllYoutubeMetadata(urlString)
	if err != nil {
		log.Printf("error: %v when checking if url '%s' is available\n", err, urlString)
	}
	return err == nil
}

func convertToAudio(videoFile string, youtubeMetadata *youtube.Video, path string) (string, error) {
	fileName := filepath.Base(videoFile)
	fileNameWithoutExtension := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	audioFileName := fmt.Sprintf("%s.mp3", fileNameWithoutExtension)
	tempAudioFilePath := filepath.Join(os.TempDir(), audioFileName)

	err := convertvideo.ConvertVideoToAudio(videoFile, tempAudioFilePath)
	if err != nil {
		return "", err
	}
	metadata := getAudioMetaData(youtubeMetadata)
	err = mp3joiner.SetFFmpegMetadataTag(tempAudioFilePath, metadata, make([]mp3joiner.Chapter, 0))
	if err != nil {
		return "", err
	}
	log.Printf("converted video file '%s' to audio '%s'\n", youtubeMetadata.Title, audioFileName)

	audioFilePath := filepath.Join(path, audioFileName)
	log.Printf("moving audio file '%s' to '%s'\n", tempAudioFilePath, audioFilePath)
	err = filemanagement.MoveFile(tempAudioFilePath, audioFilePath)
	if err != nil {
		return "", err
	}

	return audioFilePath, nil
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

	description := strings.ReplaceAll(youtubeMetadata.Description, "\n", "`n")
	description = strings.ReplaceAll(description, "\r", "`r")
	result[PodcastDescriptionTag] = description

	return result
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
	sanitizedName := sanitizeFilename(videoName)
	fileName := fmt.Sprintf("%s.mp4", sanitizedName)
	filePath := filepath.Join(path, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	log.Printf("downloading '%s' to '%s'\n", videoName, filePath)
	_, err = io.Copy(file, stream)
	if err != nil {
		return "", err
	}

	return filePath, nil
}

func getVideoId(urlString string) (id string, err error) {
	id, err = getId(urlString, videoRegex, "could not find video id in url '%s'")
	if err != nil {
		id, err = getId(urlString, videoShortRegex, "could not find video id in url '%s'")
	}
	return id, err
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
