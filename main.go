package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server"
)

func getConfigPath() string {
	// First check if config path is provided via environment variable
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}
	
	// Default to config/config.yaml in current working directory
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	return filepath.Join(cwd, "config", "config.yaml")
}

func main() {
	// Load configuration
	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Printf("Failed to load config from %s: %v", configPath, err)
		panic(err)
	}

	// Initialize database
	databaseService, err := database.NewDatabase(cfg.Persistence.Database.ConnectionString, cfg.Persistence.Media.MediaPath)
	if err != nil {
		panic(err)
	}

	// Start server
	server.StartServer(databaseService, cfg)
}
