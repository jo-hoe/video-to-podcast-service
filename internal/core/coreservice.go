package core

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
)

type CoreService struct {
	databaseService      database.DatabaseService
	audioSourceDirectory string
	cookiesConfig        *config.Cookies
}

func NewCoreService(databaseService database.DatabaseService, audioSourceDirectory string, cookiesConfig *config.Cookies) *CoreService {
	return &CoreService{
		databaseService:      databaseService,
		audioSourceDirectory: audioSourceDirectory,
		cookiesConfig:        cookiesConfig,
	}
}

func (cs *CoreService) GetDatabaseService() database.DatabaseService {
	return cs.databaseService
}

func (cs *CoreService) GetAudioSourceDirectory() string {
	return cs.audioSourceDirectory
}

func (cs *CoreService) GetLinkToFeed(host string, apiPath string, audioFilePath string) string {
	pathWithoutRoot := cs.getPathWithoutRoot(audioFilePath)
	// get first part of the path as feed title
	parts := strings.Split(pathWithoutRoot, string(os.PathSeparator))
	if len(parts) == 0 {
		log.Printf("error: audio file path '%s' does not contain a valid feed title", audioFilePath)
		return ""
	}
	feedTitle := parts[0]
	// URL encode the feed title
	urlEncodedFeedTitle := url.PathEscape(feedTitle)

	return fmt.Sprintf("http://%s/%s/%s/rss.xml", host, apiPath, urlEncodedFeedTitle)
}

func (cs *CoreService) GetLinkToAudioFile(host string, apiPath string, audioFilePath string) string {
	pathWithoutRoot := cs.getPathWithoutRoot(audioFilePath)
	parts := strings.Split(pathWithoutRoot, string(os.PathSeparator))
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	audioUrlPath := strings.Join(parts, "/")

	return fmt.Sprintf("http://%s/%s/%s", host, apiPath, audioUrlPath)
}

func (cs *CoreService) getPathWithoutRoot(audioFilePath string) string {
	pathWithoutRoot := strings.TrimPrefix(audioFilePath, cs.audioSourceDirectory)
	pathWithoutRoot = strings.TrimPrefix(pathWithoutRoot, string(os.PathSeparator))
	return pathWithoutRoot
}

func (cs *CoreService) DownloadItemsHandler(url string) (err error) {
	downloaderInstance, err := download.GetVideoDownloader(url, cs.cookiesConfig)
	if err != nil {
		return err
	}
	if !downloaderInstance.IsVideoAvailable(url) {
		return fmt.Errorf("video %s is not available", url)
	}
	log.Printf("downloading '%s'", url)

	// Refactored: move download logic to a helper function to reduce nesting and improve error handling
	go cs.handleDownload(url, downloaderInstance)

	return nil
}

func (cs *CoreService) GetFeedDirectory(audioFilePath string) (string, error) {
	if audioFilePath == "" {
		return "", fmt.Errorf("audio file path is empty")
	}
	pathWithoutRoot := cs.getPathWithoutRoot(audioFilePath)
	parts := strings.Split(pathWithoutRoot, string(os.PathSeparator))
	if len(parts) < 2 {
		return "", fmt.Errorf("audio file path '%s' does not contain a valid parent folder", audioFilePath)
	}
	parentFolder := strings.Join(parts[:len(parts)-1], string(os.PathSeparator))
	return parentFolder, nil
}

// handleDownload performs the download and podcast item creation with improved error handling and less nesting
func (cs *CoreService) handleDownload(url string, downloader downloader.AudioDownloader) {
	const maxErrorCount = 4

	filePaths, err := downloader.Download(url, cs.audioSourceDirectory)
	if err != nil {
		log.Printf("failed to download '%s': %v", url, err)
		return
	}

	for _, filePath := range filePaths {
		retries := 0
		for retries < maxErrorCount {
			podcastItem, err := database.NewPodcastItem(filePath)
			if err != nil {
				log.Printf("failed to create podcast item for '%s': %v", filePath, err)
				retries++
				continue
			}

			err = cs.databaseService.CreatePodcastItem(podcastItem)
			if err != nil {
				log.Printf("failed to create podcast item for '%s': %v", filePath, err)
				retries++
				continue
			}

			log.Printf("successfully created podcast item for '%s'", filePath)
			break // success, move to next file
		}
		if retries == maxErrorCount {
			log.Printf("giving up on '%s' after %d attempts", filePath, maxErrorCount)
		}
	}
}
