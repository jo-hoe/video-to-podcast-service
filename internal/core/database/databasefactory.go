package database

import (
	"fmt"
	"log"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/filemanagement"
)

func NewDatabase(connectionString string, resourcePath string) (database DatabaseService, err error) {
	cfg := config.GetConfig()
	var dbInstance DatabaseService

	switch cfg.Persistence.Database.Driver {
	case "sqlite3", "sqlite":
		dbInstance = NewSQLiteDatabase(connectionString)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Persistence.Database.Driver)
	}

	if !dbInstance.DoesDatabaseExist() {
		log.Print("database does not exist, creating new database")
		_, err = dbInstance.CreateDatabase()
		if err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
	} else {
		log.Print("database already exists, skipping creation.")
		_, err = dbInstance.InitializeDatabase()
		if err != nil {
			return nil, fmt.Errorf("failed to initialize database: %w", err)
		}
	}

	addPreexistingElements(dbInstance, resourcePath)

	return dbInstance, nil
}

func addPreexistingElements(database DatabaseService, resourcePath string) {
	if database == nil {
		return
	}

	log.Print("discovering preexisting audio files in resource path: ", resourcePath)
	filePaths, err := filemanagement.GetAudioFiles(resourcePath)

	if err != nil {
		fmt.Printf("error initializing database while retrieving audio files: %v\n", err)
		return
	}

	for _, file := range filePaths {
		podcastItem, err := NewPodcastItem(file)
		if err != nil {
			fmt.Printf("error creating podcast item for file %s: %v\n", file, err)
			continue
		}

		err = database.CreatePodcastItem(podcastItem)
		if err != nil {
			fmt.Printf("error saving podcast item for file %s: %v\n", file, err)
		}
	}
	log.Printf("added %d preexisting podcast items to the database", len(filePaths))
}
