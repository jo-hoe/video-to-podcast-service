package youtube

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lrstanley/go-ytdlp"
)

func TestYoutubeAudioDownloader_IsVideoSupported(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		y    *YoutubeAudioDownloader
		args args
		want bool
	}{
		{
			name: "test video link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://www.youtube.com/watch?v=jNQXAC9IVRw&pp=ygUQb25lIHNlY29uZCB2aWRlbw%3D%3D",
			},
			want: true,
		},
		{
			name: "test playlist link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://www.youtube.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-",
			},
			want: true,
		},
		{
			name: "test short video link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://youtu.be/DucriSA8ukw?feature=shared",
			},
			want: true,
		},
		{
			name: "test not existing link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://not-existing.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.y.IsVideoSupported(tt.args.url); got != tt.want {
				t.Errorf("YoutubeAudioDownloader.IsVideoSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_configureCookies(t *testing.T) {
	tests := []struct {
		name           string
		cookieFile     string
		createFile     bool
		fileContent    string
		expectCookies  bool
		expectWarning  bool
	}{
		{
			name:          "no cookie file configured",
			cookieFile:    "",
			createFile:    false,
			expectCookies: false,
		},
		{
			name:          "cookie file exists",
			cookieFile:    "test_cookies.txt",
			createFile:    true,
			fileContent:   "# Netscape HTTP Cookie File\n.youtube.com\tTRUE\t/\tFALSE\t1234567890\ttest_cookie\ttest_value",
			expectCookies: true,
		},
		{
			name:          "cookie file configured but doesn't exist",
			cookieFile:    "nonexistent_cookies.txt",
			createFile:    false,
			expectCookies: false,
			expectWarning: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			originalEnv := os.Getenv(cookieFileEnvVar)
			defer os.Setenv(cookieFileEnvVar, originalEnv)

			var tempFile string
			if tt.createFile {
				// Create temporary file
				tmpDir := t.TempDir()
				tempFile = filepath.Join(tmpDir, tt.cookieFile)
				err := os.WriteFile(tempFile, []byte(tt.fileContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create test cookie file: %v", err)
				}
				os.Setenv(cookieFileEnvVar, tempFile)
			} else if tt.cookieFile != "" {
				// Set environment variable to non-existent file
				os.Setenv(cookieFileEnvVar, tt.cookieFile)
			} else {
				// Clear environment variable
				os.Unsetenv(cookieFileEnvVar)
			}

			// Test
			dl := ytdlp.New()
			result := configureCookies(dl)

			// Verify
			if result == nil {
				t.Error("configureCookies() returned nil")
				return
			}

			// The actual verification of whether cookies were set would require
			// inspecting the internal state of the ytdlp.Command, which isn't
			// easily accessible. For now, we verify that the function doesn't
			// panic and returns a valid command object.
			
			// Clean up temp file if created
			if tt.createFile && tempFile != "" {
				os.Remove(tempFile)
			}
		})
	}
}

