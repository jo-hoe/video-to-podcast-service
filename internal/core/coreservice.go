package core

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
)

const (
	baseURLEnvVar = "BASE_URL"
)

type CoreService struct {
	databaseService      database.DatabaseService
	audioSourceDirectory string
}

func NewCoreService(databaseService database.DatabaseService, audioSourceDirectory string) *CoreService {
	return &CoreService{
		databaseService:      databaseService,
		audioSourceDirectory: audioSourceDirectory,
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
	// Normalize path separators to forward slashes for URLs
	pathWithoutRoot = strings.ReplaceAll(pathWithoutRoot, string(os.PathSeparator), "/")
	// get first part of the path as feed title
	parts := strings.Split(pathWithoutRoot, "/")
	if len(parts) == 0 {
		log.Printf("error: audio file path '%s' does not contain a valid feed title", audioFilePath)
		return ""
	}
	feedTitle := parts[0]
	// URL encode the feed title
	urlEncodedFeedTitle := url.PathEscape(feedTitle)

	baseURL := cs.getBaseURL(host)
	return fmt.Sprintf("%s/%s/%s/rss.xml", baseURL, apiPath, urlEncodedFeedTitle)
}

func (cs *CoreService) GetLinkToAudioFile(host string, apiPath string, audioFilePath string) string {
	pathWithoutRoot := cs.getPathWithoutRoot(audioFilePath)
	// Normalize path separators to forward slashes for URLs
	pathWithoutRoot = strings.ReplaceAll(pathWithoutRoot, string(os.PathSeparator), "/")
	parts := strings.Split(pathWithoutRoot, "/")
	for i, part := range parts {
		parts[i] = url.PathEscape(part)
	}
	audioUrlPath := strings.Join(parts, "/")

	baseURL := cs.getBaseURL(host)
	return fmt.Sprintf("%s/%s/%s", baseURL, apiPath, audioUrlPath)
}

func (cs *CoreService) getPathWithoutRoot(audioFilePath string) string {
	pathWithoutRoot := strings.TrimPrefix(audioFilePath, cs.audioSourceDirectory)
	// Remove all leading slashes/backslashes (Windows/Linux)
	pathWithoutRoot = strings.TrimLeft(pathWithoutRoot, "\\/")
	return pathWithoutRoot
}

// getBaseURL returns the base URL for generating links, prioritizing environment variable over host header
func (cs *CoreService) getBaseURL(host string) string {
	if baseURL := os.Getenv(baseURLEnvVar); baseURL != "" {
		// Remove trailing slash if present
		baseURL = strings.TrimSuffix(baseURL, "/")
		// Ensure the URL has a proper scheme
		if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
			// Default to https if no scheme is provided
			baseURL = "https://" + baseURL
		}
		return baseURL
	}

	// Fallback to host header with http scheme (existing behavior)
	return "http://" + host
}

func (cs *CoreService) DownloadItemsHandler(url string) (err error) {
	downloaderInstance, err := download.GetVideoDownloader(url)
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
