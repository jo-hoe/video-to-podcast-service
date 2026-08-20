package core

import (
	"net/url"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
)

// MockService is a test double for Service. Override fields to inject specific behaviour;
// zero values produce safe no-op defaults.
type MockService struct {
	DatabaseService         database.DatabaseService
	AudioSourceDirectory    string
	CookieConfig            *config.Cookies
	DownloadItemsHandlerFunc func(url string) error
	DeletePodcastItemFunc   func(id string) error
	GetFeedDirectoryFunc    func(audioFilePath string) (string, error)
}

func NewMockService() *MockService {
	return &MockService{
		DatabaseService: database.NewMockDatabase(),
	}
}

func (m *MockService) GetDatabaseService() database.DatabaseService {
	return m.DatabaseService
}

func (m *MockService) GetAudioSourceDirectory() string {
	return m.AudioSourceDirectory
}

func (m *MockService) GetCookieConfig() *config.Cookies {
	return m.CookieConfig
}

func (m *MockService) GetFeedDirectory(audioFilePath string) (string, error) {
	if m.GetFeedDirectoryFunc != nil {
		return m.GetFeedDirectoryFunc(audioFilePath)
	}
	return "", nil
}

func (m *MockService) GetLinkToFeed(_ *url.URL, _ string, _ string) string {
	return ""
}

func (m *MockService) GetLinkToAudioFile(_ *url.URL, _ string, _ string) string {
	return ""
}

func (m *MockService) DeletePodcastItem(id string) error {
	if m.DeletePodcastItemFunc != nil {
		return m.DeletePodcastItemFunc(id)
	}
	return nil
}

func (m *MockService) DownloadItemsHandler(url string) error {
	if m.DownloadItemsHandlerFunc != nil {
		return m.DownloadItemsHandlerFunc(url)
	}
	return nil
}
