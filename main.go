package main

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

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
	// Initialize a temporary default slog logger at INFO level
	tmpHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	slog.SetDefault(slog.New(tmpHandler))

	// Load configuration
	configPath := getConfigPath()
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		slog.Error("Failed to load config", "path", configPath, "err", err)
		panic(err)
	}

	// Reconfigure slog according to configured log level
	level := parseLogLevel(cfg.LogLevel)
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))

	// Initialize database
	databaseService, err := database.NewDatabase(cfg.Persistence.Database.ConnectionString, cfg.Persistence.Media.MediaPath)
	if err != nil {
		slog.Error("Failed to initialize database", "err", err)
		panic(err)
	}

	// Start server
	server.StartServer(databaseService, cfg)
}

// parseLogLevel maps a string to slog.Level with a safe default.
func parseLogLevel(lvl string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(lvl)) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
