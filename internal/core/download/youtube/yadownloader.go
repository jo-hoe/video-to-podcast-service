package youtube

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
	"github.com/lrstanley/go-ytdlp"
	"golang.org/x/net/context"
)

const (
	playlistRegex   = `https://(?:.+)?youtube.com/(?:.+)?list=([A-Za-z0-9_-]*)`
	videoRegex      = `https://(?:.+)?youtube.com/(?:.+)?watch\?v=([A-Za-z0-9_-]*)`
	videoShortRegex = `https://youtu\.be/([A-Za-z0-9_-]*)`
	// types taken from API description
	// https://wiki.sponsor.ajay.app/w/Types
	sponsorBlockCategories = "sponsor,selfpromo,interaction,intro,outro,preview,music_offtopic,filler"
	// ID3 tag youtube-dl uses to store the video URL
	VideoUrlID3KeyAttribute = "purl"
)

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
	defer func() {
		if err := os.RemoveAll(tempPath); err != nil {
			log.Printf("error removing temp directory: %v", err)
		}
	}()

	log.Printf("downloading from '%s' to '%s'", urlString, tempPath)
	tempResults, err := download(tempPath, urlString)
	if err != nil {
		return nil, err
	}
	log.Printf("done downloading %d files", len(tempResults))

	for _, filePath := range tempResults {
		log.Printf("setting metadata for '%s'", filePath)
		err = setMetadata(filePath)
		if err != nil {
			return nil, err
		}
		log.Printf("set metadata for '%s'", filePath)
	}

	log.Printf("moving files to target folder")
	for _, filePath := range tempResults {
		movedItem, err := moveToTarget(filePath, targetPath)
		if err != nil {
			return results, err
		}
		results = append(results, movedItem)
	}
	log.Printf("completed moving all relevant files")

	return results, err
}

func setMetadata(fullFilePath string) (err error) {
	metadata, err := mp3joiner.GetFFmpegMetadataTag(fullFilePath)
	if err != nil {
		return err
	}
	chapters, err := mp3joiner.GetChapterMetadata(fullFilePath)
	if err != nil {
		return err
	}

	metadata[downloader.PodcastDescriptionTag] = strings.ReplaceAll(metadata["synopsis"], "\n", "<br>")
	metadata[downloader.DateTag] = metadata["date"]
	metadata[downloader.VideoDownloadLink] = metadata[VideoUrlID3KeyAttribute]

	videoUrl := metadata["purl"]
	thumbnailUrl, err := getThumbnailUrl(videoUrl)
	if err != nil {
		return err
	}
	metadata[downloader.ThumbnailUrlTag] = thumbnailUrl

	return mp3joiner.SetFFmpegMetadataTag(fullFilePath, metadata, chapters)
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

func moveToTarget(sourcePath, targetRootPath string) (results string, err error) {
	// move file to target directory
	// the target directory is created based on the source file's parent directory
	//
	// e.g.:
	// sourcePath = /tmp/1234/5678/file.mp3
	// targetRootPath = /podcasts
	// results = /podcasts/5678/file.mp3
	directoryName := filepath.Base(filepath.Dir(sourcePath))
	targetSubDirectory := filepath.Join(targetRootPath, directoryName)
	err = os.MkdirAll(targetSubDirectory, os.ModePerm)
	if err != nil {
		return results, err
	}

	targetFilename := filepath.Base(sourcePath)
	targetPath := filepath.Join(targetSubDirectory, targetFilename)
	err = filemanagement.MoveFile(sourcePath, targetPath)
	if err != nil {
		return results, err
	}
	return targetPath, err
}

func download(targetDirectory string, urlString string) ([]string, error) {
	result := make([]string, 0)
	// set download behavior
	tempFilenameTemplate := fmt.Sprintf("%s%c%s", targetDirectory, os.PathSeparator, "%(channel)s/%(title)s_%(id)s.%(ext)s")
	dl := ytdlp.New().
		ExtractAudio().AudioFormat("mp3").          // convert get mp3 after downloading the video
		EmbedMetadata().                            // adds metadata such as artist to the file
		SponsorblockRemove(sponsorBlockCategories). // delete unneeded segments (e.g. sponsor, intro etc.)
		ProgressFunc(1*time.Second, func(prog ytdlp.ProgressUpdate) {
			log.Printf("download progress '%s' - %.1f%%", *prog.Info.Title, prog.Percent())
		}).
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
	log.Printf("checking if video from '%s' can be downloaded", urlString)
	dl, err := ytdlp.New().Simulate().Quiet().Run(context.Background(), urlString)
	if err != nil {
		log.Printf("error checking video availability: '%v'", err)
		return false
	}
	return dl != nil
}
