package download

import (
	"os"
	"path/filepath"
	"testing"
)

const (
	validYoutubeVideoUrl    = "https://www.youtube.com/watch?v=jNQXAC9IVRw"
	validYoutubePlaylistUrl = "https://www.youtube.com/playlist?list=PLHJH2BlYG-EEBtw2y1njWpDukJSTs8Qqx"
)

// Skips integration test if requirements are not meet
func checkPrerequisites(t *testing.T) {
	// Some servers/IPs are blocked by Youtube and the test will fail
	// this includes Github Actions servers
	if os.Getenv("GITHUB_ACTIONS") == "true" {
		t.Skip("Test will be skipped in Github Context")
	}
}

func Test_YoutubeAudioDownloader_Download(t *testing.T) {
	checkPrerequisites(t)

	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	defer os.RemoveAll(rootDirectory)
	if err != nil {
		t.Error("could not create folder")
	}

	type args struct {
		urlString string
		path      string
	}
	tests := []struct {
		name    string
		y       *YoutubeAudioDownloader
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Video Download Test",
			y:    &YoutubeAudioDownloader{},
			args: args{
				urlString: validYoutubeVideoUrl,
				path:      rootDirectory,
			},
			want:    []string{filepath.Join(rootDirectory, "kids", "kids video.mp3")},
			wantErr: false,
		},
		{
			name: "Playlist Download Test",
			y:    &YoutubeAudioDownloader{},
			args: args{
				urlString: validYoutubePlaylistUrl,
				path:      filepath.Join(rootDirectory, "Cat"),
			},
			want:    make([]string, 10),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.y.Download(tt.args.urlString, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoutubeAudioDownloader.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("YoutubeAudioDownloader.Download() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYoutubeAudioDownloader_IsVideoAvailable(t *testing.T) {
	checkPrerequisites(t)
	
	type args struct {
		urlString string
	}
	tests := []struct {
		name string
		y    *YoutubeAudioDownloader
		args args
		want bool
	}{
		{
			name: "Check valid url",
			y:    &YoutubeAudioDownloader{},
			args: args{
				urlString: validYoutubeVideoUrl,
			},
			want: true,
		},
		{
			name: "Check invalid url",
			y:    &YoutubeAudioDownloader{},
			args: args{
				urlString: "https://www.youtube.com/watch?v=invalid_url",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.y.IsVideoAvailable(tt.args.urlString); got != tt.want {
				t.Errorf("YoutubeAudioDownloader.IsVideoAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
