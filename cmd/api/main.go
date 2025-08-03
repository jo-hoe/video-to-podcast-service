package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-playground/validator"
	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server/api"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	configPath := config.GetConfigPath()

	// Create default config if it doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Printf("Config file not found at %s, creating default configuration", configPath)
		if err := config.SaveDefaultConfig(configPath); err != nil {
			log.Printf("Warning: Failed to create default config file: %v", err)
		}
	}

	cfg, err := config.LoadAPIConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Log configuration
	log.Printf("Starting API service on port %s", cfg.Server.Port)
	log.Printf("Database: %s", cfg.Database.ConnectionString)
	log.Printf("Storage path: %s", cfg.Storage.BasePath)
	if cfg.External.YTDLPCookiesFile != "" {
		log.Printf("Using cookies file: %s", cfg.External.YTDLPCookiesFile)
	}

	// Ensure directory structure exists
	if err := ensureDirectoryStructure(cfg); err != nil {
		log.Fatalf("Failed to create directory structure: %v", err)
	}

	databaseService, err := database.NewDatabase(cfg.Database.ConnectionString, cfg.Storage.BasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	startAPIServer(cfg, databaseService)
}

func startAPIServer(cfg *config.APIConfig, databaseService database.DatabaseService) {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.Validator = &genericValidator{Validator: validator.New()}

	coreService := core.NewCoreService(databaseService, cfg.Storage.BasePath)

	apiService := api.NewAPIService(coreService, cfg.Server.Port)
	apiService.SetAPIRoutes(e)

	// start API server
	log.Print("starting API server")
	log.Printf("API available at http://localhost:%s/%s", cfg.Server.Port, api.FeedsPath)
	log.Printf("Health check available at http://localhost:%s/v1/health", cfg.Server.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf(":%s", cfg.Server.Port)))
}

type genericValidator struct {
	Validator *validator.Validate
}

// ensureDirectoryStructure creates the necessary directory structure for the application
func ensureDirectoryStructure(cfg *config.APIConfig) error {
	// Create resources directory
	if err := os.MkdirAll(cfg.Storage.BasePath, 0755); err != nil {
		return fmt.Errorf("failed to create resources directory: %w", err)
	}

	// Create database directory if using file-based database
	if cfg.Database.ConnectionString != "" {
		dbDir := filepath.Dir(cfg.Database.ConnectionString)
		if dbDir != "." && dbDir != "" {
			if err := os.MkdirAll(dbDir, 0755); err != nil {
				return fmt.Errorf("failed to create database directory: %w", err)
			}
		}
	}

	// Create cookies directory if cookies file is configured
	if cfg.External.YTDLPCookiesFile != "" {
		cookiesDir := filepath.Dir(cfg.External.YTDLPCookiesFile)
		if cookiesDir != "." && cookiesDir != "" {
			if err := os.MkdirAll(cookiesDir, 0755); err != nil {
				return fmt.Errorf("failed to create cookies directory: %w", err)
			}
		}
	}

	return nil
}

func (gv *genericValidator) Validate(i interface{}) error {
	if err := gv.Validator.Struct(i); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("received invalid request body: %v", err))
	}
	return nil
}
