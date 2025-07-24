package core

import (
	"os"
	"testing"
)

func TestGetBaseURL(t *testing.T) {
	cs := &CoreService{}

	tests := []struct {
		name     string
		envValue string
		host     string
		expected string
	}{
		{
			name:     "no env var, uses host with http",
			envValue: "",
			host:     "localhost:8080",
			expected: "http://localhost:8080",
		},
		{
			name:     "env var with https scheme",
			envValue: "https://example.com",
			host:     "localhost:8080",
			expected: "https://example.com",
		},
		{
			name:     "env var with http scheme",
			envValue: "http://example.com:8080",
			host:     "localhost:8080",
			expected: "http://example.com:8080",
		},
		{
			name:     "env var without scheme defaults to https",
			envValue: "example.com",
			host:     "localhost:8080",
			expected: "https://example.com",
		},
		{
			name:     "env var with trailing slash removed",
			envValue: "https://example.com/",
			host:     "localhost:8080",
			expected: "https://example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			originalEnv := os.Getenv(baseURLEnvVar)
			defer func() {
				_ = os.Setenv(baseURLEnvVar, originalEnv)
			}()

			if tt.envValue != "" {
				_ = os.Setenv(baseURLEnvVar, tt.envValue)
			} else {
				_ = os.Unsetenv(baseURLEnvVar)
			}

			result := cs.getBaseURL(tt.host)
			if result != tt.expected {
				t.Errorf("getBaseURL() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
func TestGetLinkToFeed(t *testing.T) {
	cs := &CoreService{audioSourceDirectory: "/app/resources"}

	tests := []struct {
		name          string
		baseURLEnv    string
		host          string
		apiPath       string
		audioFilePath string
		expected      string
	}{
		{
			name:          "with BASE_URL env var",
			baseURLEnv:    "https://podcasts.example.com",
			host:          "localhost:8080",
			apiPath:       "v1/feeds",
			audioFilePath: "/app/resources/channel1/audio.mp3",
			expected:      "https://podcasts.example.com/v1/feeds/channel1/rss.xml",
		},
		{
			name:          "without BASE_URL env var, uses host",
			baseURLEnv:    "",
			host:          "192.168.1.100:8080",
			apiPath:       "v1/feeds",
			audioFilePath: "/app/resources/my-channel/audio.mp3",
			expected:      "http://192.168.1.100:8080/v1/feeds/my-channel/rss.xml",
		},
		{
			name:          "URL encoding of feed title",
			baseURLEnv:    "https://example.com",
			host:          "localhost:8080",
			apiPath:       "v1/feeds",
			audioFilePath: "/app/resources/channel with spaces/audio.mp3",
			expected:      "https://example.com/v1/feeds/channel%20with%20spaces/rss.xml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			originalEnv := os.Getenv(baseURLEnvVar)
			defer func() {
				_ = os.Setenv(baseURLEnvVar, originalEnv)
			}()

			if tt.baseURLEnv != "" {
				_ = os.Setenv(baseURLEnvVar, tt.baseURLEnv)
			} else {
				_ = os.Unsetenv(baseURLEnvVar)
			}

			result := cs.GetLinkToFeed(tt.host, tt.apiPath, tt.audioFilePath)
			if result != tt.expected {
				t.Errorf("GetLinkToFeed() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestGetLinkToAudioFile(t *testing.T) {
	cs := &CoreService{audioSourceDirectory: "/app/resources"}

	tests := []struct {
		name          string
		baseURLEnv    string
		host          string
		apiPath       string
		audioFilePath string
		expected      string
	}{
		{
			name:          "with BASE_URL env var",
			baseURLEnv:    "https://podcasts.example.com",
			host:          "localhost:8080",
			apiPath:       "v1/audio",
			audioFilePath: "/app/resources/channel1/my-audio.mp3",
			expected:      "https://podcasts.example.com/v1/audio/channel1/my-audio.mp3",
		},
		{
			name:          "without BASE_URL env var, uses host",
			baseURLEnv:    "",
			host:          "192.168.1.100:8080",
			apiPath:       "v1/audio",
			audioFilePath: "/app/resources/my-channel/audio file.mp3",
			expected:      "http://192.168.1.100:8080/v1/audio/my-channel/audio%20file.mp3",
		},
		{
			name:          "URL encoding of file path",
			baseURLEnv:    "https://example.com",
			host:          "localhost:8080",
			apiPath:       "v1/audio",
			audioFilePath: "/app/resources/channel with spaces/file with spaces.mp3",
			expected:      "https://example.com/v1/audio/channel%20with%20spaces/file%20with%20spaces.mp3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			originalEnv := os.Getenv(baseURLEnvVar)
			defer func() {
				_ = os.Setenv(baseURLEnvVar, originalEnv)
			}()

			if tt.baseURLEnv != "" {
				_ = os.Setenv(baseURLEnvVar, tt.baseURLEnv)
			} else {
				_ = os.Unsetenv(baseURLEnvVar)
			}

			result := cs.GetLinkToAudioFile(tt.host, tt.apiPath, tt.audioFilePath)
			if result != tt.expected {
				t.Errorf("GetLinkToAudioFile() = %v, expected %v", result, tt.expected)
			}
		})
	}
}
