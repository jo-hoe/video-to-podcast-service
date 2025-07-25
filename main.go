package main

import (
	"log"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server"
)

func main() {
	cfg := config.LoadConfig()

	databaseService, err := database.NewDatabase(cfg.Database.ConnectionString, cfg.Storage.BasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	server.StartServer(cfg, databaseService)
}
