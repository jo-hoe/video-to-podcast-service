package download

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
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

func (y *YoutubeAudioDownloader) Download(urlString string, targetPath string) ([]string, error) {
	results := make([]string, 0)
	// create temp directory
	tempPath, err := os.MkdirTemp("", "")
	if err != nil {
		return results, err
	}
	defer os.RemoveAll(tempPath)

	log.Printf("downloading from '%s' to '%s'", urlString, tempPath)
	tempResults, err := download(tempPath, urlString)
	if err != nil {
		return nil, err
	}
	log.Printf("done downloading %d files", len(tempResults))

	log.Printf("setting metadata")
	err = setMetadata(tempResults)
	if err != nil {
		return nil, err
	}
	log.Printf("set all metadata")

	log.Printf("moving files to target folder")
	results, err = moveToTarget(tempResults, targetPath)
	if err != nil {
		return results, err
	}
	log.Printf("completed moving all relevant files")

	return results, err
}

func setMetadata(tempResults []string) (err error) {
	for _, fullFilePath := range tempResults {
		metadata, err := mp3joiner.GetFFmpegMetadataTag(fullFilePath)
		if err != nil {
			return err
		}
		chapters, err := mp3joiner.GetChapterMetadata(fullFilePath)
		if err != nil {
			return err
		}
		metadata[PodcastDescriptionTag] = metadata["description"]
		thumbnailUrl, err := getThumbnailUrl(metadata["purl"])
		if err != nil {
			return err
		}
		metadata[ThumbnailUrlTag] = thumbnailUrl
		mp3joiner.SetFFmpegMetadataTag(fullFilePath, metadata, chapters)
	}

	return err
}

func getThumbnailUrl(videoUrl string) (result string, err error) {
	dl := ytdlp.New().GetThumbnail()

	cliOutput, err := dl.Run(context.Background(), videoUrl)
	if err != nil {
		log.Printf("error getting thumbnail rul from '%s': '%v'", videoUrl, err)
		return result, err
	}
	for _, output := range cliOutput.OutputLogs {
		if strings.HasPrefix(output.Line, "https") {
			result = output.Line
		}
	}

	return result, err
}

func moveToTarget(tempResults []string, targetPath string) (results []string, err error) {
	results = make([]string, 0)
	for _, fullFilePath := range tempResults {
		directoryName := filepath.Base(filepath.Dir(fullFilePath))
		targetSubDirectory := filepath.Join(targetPath, directoryName)
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
		ParseMetadata("description:TDES").
		Output(tempFilenameTemplate) // set output path

	// download
	_, err := dl.Run(context.Background(), urlString)
	if err != nil {
		return result, err
	}

	// get file names
	result, err = filemanagement.GetAudioFiles(targetDirectory)
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
