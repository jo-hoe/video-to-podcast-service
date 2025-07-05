package database

import (
	"fmt"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
)

func NewDatabase(dbType string, dataSourceName string, resourcePath string) (database Database, err error) {
	switch dbType {
	case "sqlite":
		database, err = NewSQLiteDatabase(dataSourceName)
	default:
		err = fmt.Errorf("unsupported database type: %s", dbType)
	}

	if err != nil {
		initializeDatabase(database, resourcePath)
	}

	return database, err
}

func initializeDatabase(database Database, resourcePath string) {
	if database == nil {
		return
	}

	filePaths, err := filemanagement.GetAudioFiles(resourcePath)

	if err != nil {
		fmt.Printf("Error initializing database while retrieving audio files: %v\n", err)
		return
	}

	for _, file := range filePaths {
		podcastItem, err := NewPodcastItem(file)
		if err != nil {
			fmt.Printf("Error creating podcast item for file %s: %v\n", file, err)
			continue
		}

		err = database.CreatePodcastItem(podcastItem)
		if err != nil {
			fmt.Printf("Error saving podcast item for file %s: %v\n", file, err)
		}
	}
}
