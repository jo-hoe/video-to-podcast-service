package core

import (
	"net/url"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
)

// Service is the interface that the API layer depends on.
type Service interface {
	GetDatabaseService() database.DatabaseService
	GetAudioSourceDirectory() string
	GetCookieConfig() *config.Cookies
	GetFeedDirectory(audioFilePath string) (string, error)
	GetLinkToFeed(baseURL *url.URL, apiPath string, audioFilePath string) string
	GetLinkToAudioFile(baseURL *url.URL, apiPath string, audioFilePath string) string
	DeletePodcastItem(id string) error
	DownloadItemsHandler(url string) error
}
