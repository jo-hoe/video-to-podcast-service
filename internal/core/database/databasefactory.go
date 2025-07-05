package database

import (
	"fmt"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
)

func NewDatabase(connectionString string, resourcePath string) (database Database, err error) {
	var dbInstance Database
	dbType := getDatabaseType(connectionString)

	switch dbType {
	case "sqlite":
		dbInstance = &SQLiteDatabase{}
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}

	if !dbInstance.DoesDatabaseExist() {
		_, err = dbInstance.CreateDatabase(connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	} else {
		_, err = dbInstance.InitializeDatabase(connectionString)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	addPreexistingElements(dbInstance, resourcePath)

	return dbInstance, nil
}

func getDatabaseType(connectionString string) string {
	if connectionString == "" {
		return "sqlite"
	}

	if len(connectionString) > 7 && connectionString[:7] == "sqlite:" {
		return "sqlite"
	}

	// Add more database types as needed
	return "unknown"
}

func addPreexistingElements(database Database, resourcePath string) {
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
