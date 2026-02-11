package core

import (
	"fmt"
	"log/slog"
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
	mediaConfig          *config.Media
}

func NewCoreService(databaseService database.DatabaseService, audioSourceDirectory string, cookiesConfig *config.Cookies, mediaConfig *config.Media) *CoreService {
	return &CoreService{
		databaseService:      databaseService,
		audioSourceDirectory: audioSourceDirectory,
		cookiesConfig:        cookiesConfig,
		mediaConfig:          mediaConfig,
	}
}

func (cs *CoreService) GetDatabaseService() database.DatabaseService {
	return cs.databaseService
}

func (cs *CoreService) GetAudioSourceDirectory() string {
	return cs.audioSourceDirectory
}

func (cs *CoreService) GetCookieConfig() *config.Cookies {
	return cs.cookiesConfig
}

func (cs *CoreService) GetLinkToFeed(host string, apiPath string, audioFilePath string) string {
	pathWithoutRoot := cs.getPathWithoutRoot(audioFilePath)
	// get first part of the path as feed title
	parts := strings.Split(pathWithoutRoot, string(os.PathSeparator))
	if len(parts) == 0 {
		slog.Error("audio file path does not contain a valid feed title", "audioFilePath", audioFilePath)
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
	downloaderInstance, err := download.GetVideoDownloader(url, cs.cookiesConfig, cs.mediaConfig)
	if err != nil {
		return fmt.Errorf("url %s not supported", url)
	}

	// Get individual entries (playlist expands to multiple URLs; single video returns itself)
	entries, err := downloaderInstance.ListVideoEntries(url)
	if err != nil {
		slog.Error("failed to list video entries", "url", url, "err", err)
		return fmt.Errorf("failed to list entries for %s", url)
	}
	if len(entries) == 0 {
		return fmt.Errorf("no downloadable entries for %s", url)
	}

	slog.Info("starting downloads", "requestedUrl", url, "entryCount", len(entries))

	// Throttle parallel downloads using a semaphore based on configured max parallel downloads (default 1)
	maxParallel := cs.mediaConfig.MaxParallelDownloads
	if maxParallel <= 0 {
		maxParallel = 1
	}
	sem := make(chan struct{}, maxParallel)

	// Schedule downloads in background to avoid blocking the API response
	go func(urls []string) {
		for _, entryURL := range urls {
			if !downloaderInstance.IsVideoAvailable(entryURL) {
				slog.Warn("video is not available, skipping", "url", entryURL)
				continue
			}
			sem <- struct{}{}
			go func(u string) {
				defer func() { <-sem }()
				cs.handleDownload(u, downloaderInstance)
			}(entryURL)
		}
	}(entries)

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

	filePath, err := downloader.Download(url, cs.audioSourceDirectory)
	if err != nil {
		slog.Error("failed to download", "url", url, "err", err)
		return
	}

	retries := 0
	for retries < maxErrorCount {
		podcastItem, err := database.NewPodcastItem(filePath)
		if err != nil {
			slog.Error("failed to create podcast item", "filePath", filePath, "err", err)
			retries++
			continue
		}

		err = cs.databaseService.InsertReplacePodcastItem(podcastItem)
		if err != nil {
			slog.Error("failed to create podcast item", "filePath", filePath, "err", err)
			retries++
			continue
		}

		slog.Info("successfully created podcast item", "filePath", filePath)
		break // success
	}
	if retries == maxErrorCount {
		slog.Warn("giving up on file after max attempts", "filePath", filePath, "attempts", maxErrorCount)
	}
}
