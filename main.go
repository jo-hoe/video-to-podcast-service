package main

import (
	"os"
	"path/filepath"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
	"github.com/jo-hoe/video-to-podcast-service/internal/server"
)

const (
	resourcePathEnvVar     = "BASE_PATH"
	connectionStringEnvVar = "CONNECTION_STRING"
)

func getResourcePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	defaultResourcePath := common.ValueOrDefault(os.Getenv(resourcePathEnvVar), filepath.Join(exPath, "resources"))
	return defaultResourcePath
}

func main() {
	defaultResourcePath := getResourcePath()

	databaseService, err := database.NewDatabase(common.ValueOrDefault(os.Getenv(connectionStringEnvVar), ""), defaultResourcePath)
	if err != nil {
		panic(err)
	}

	server.StartServer(databaseService, defaultResourcePath)
}
