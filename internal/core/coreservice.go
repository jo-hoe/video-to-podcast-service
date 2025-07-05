package core

import (
	"fmt"
	"log"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download"
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
	downloader, err := download.GetVideoDownloader(url)
	if err != nil {
		return err
	}
	if !downloader.IsVideoAvailable(url) {
		return fmt.Errorf("video %s is not available", url)
	}
	log.Printf("downloading '%s'", url)

	go func() {
		maxErrorCount := 4
		errorCount := 0

		for errorCount < maxErrorCount {
			filePaths, err := downloader.Download(url, cs.audioSourceDirectory)
			if err != nil {
				log.Printf("failed to download '%s': %v", url, err)
				errorCount++
			}

			for _, filePath := range filePaths {
				podcastItem, err := database.NewPodcastItem(filePath)
				if err != nil {
					log.Printf("failed to create podcast item for '%s': %v", filePath, err)
					errorCount++
					continue
				}

				err = cs.databaseService.CreatePodcastItem(podcastItem)
				if err != nil {
					log.Printf("failed to create podcast item for '%s': %v", filePath, err)
					errorCount++
					continue
				} else {
					log.Printf("successfully created podcast item for '%s'", filePath)
					break
				}
			}
		}
	}()

	return nil
}
