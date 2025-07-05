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
			_, err := downloader.Download(url, cs.audioSourceDirectory)
			if err != nil {
				log.Printf("failed to download '%s': %v", url, err)
				errorCount++
			} else {
				break
			}
		}
	}()

	return nil
}
