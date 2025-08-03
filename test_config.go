package main

import (
	"fmt"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("API Port: %s\n", cfg.API.Server.Port)
	fmt.Printf("UI Port: %s\n", cfg.UI.Server.Port)
	fmt.Printf("Database: %s\n", cfg.API.Database.ConnectionString)
	fmt.Printf("Storage: %s\n", cfg.API.Storage.BasePath)
	fmt.Printf("UI API URL: %s\n", cfg.UI.API.BaseURL)
}
