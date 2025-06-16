package main

import (
	"os"
	"path/filepath"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/server"
)

func getResourcePath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	exPath := filepath.Dir(ex)
	defaultResourcePath := common.ValueOrDefault(os.Getenv("BASE_PATH"), filepath.Join(exPath, "resources"))
	return defaultResourcePath
}

func main() {
	defaultResourcePath := getResourcePath()

	server.StartServer(defaultResourcePath)
}
