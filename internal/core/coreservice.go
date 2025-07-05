package core

import (
	"fmt"
	"log"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
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
