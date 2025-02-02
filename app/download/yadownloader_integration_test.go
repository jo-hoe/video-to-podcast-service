package download

import (
	"os"
	"path/filepath"
	"testing"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
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

func Test_YoutubeAudioDownloader_Download_File_Properties(t *testing.T) {
	checkPrerequisites(t)

	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	defer os.RemoveAll(rootDirectory)
	if err != nil {
		t.Error("could not create folder")
	}

	y := NewYoutubeAudioDownloader()
	result, err := y.Download(validYoutubeVideoUrl, rootDirectory)
	if err != nil {
		t.Errorf("YoutubeAudioDownloader.Download() error = %v", err)
	}
	if len(result) == 0 {
		t.Errorf("YoutubeAudioDownloader.Download() = %v, want non empty", result)
	}
	if len(result) > 1 {
		t.Errorf("YoutubeAudioDownloader.Download() = %v, want only one", result)
	}

	metadata, err := mp3joiner.GetFFmpegMetadataTag(result[0])
	if err != nil {
		t.Errorf("YoutubeAudioDownloader.Download() error = %v", err)
	}
	expectedArtist := "jawed"
	if metadata["artist"] != expectedArtist {
		t.Errorf("YoutubeAudioDownloader.Download() = %v, want %v", metadata["artist"], expectedArtist)
	}

	expectedFilename := "Me at the zoo.mp3"
	if result[0] != filepath.Join(rootDirectory, expectedArtist, expectedFilename) {
		t.Errorf("YoutubeAudioDownloader.Download() = %v, want %v", result[0], filepath.Join(rootDirectory, expectedArtist, expectedFilename))
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
			y:    NewYoutubeAudioDownloader(),
			args: args{
				urlString: validYoutubeVideoUrl,
				path:      rootDirectory,
			},
			want:    []string{filepath.Join(rootDirectory, "jawed", "Me at the zoo.mp3")},
			wantErr: false,
		},
		{
			name: "Playlist Download Test",
			y:    NewYoutubeAudioDownloader(),
			args: args{
				urlString: validYoutubePlaylistUrl,
				path:      filepath.Join(rootDirectory, "Cat"),
			},
			want: []string{
				filepath.Join(rootDirectory, "Shortest Video on Youtube.mp3"),
				filepath.Join(rootDirectory, "Shortest Video on Youtube Part 2.mp3"),
				filepath.Join(rootDirectory, "Shortest Video on Youtube Part 3.mp3"),
				filepath.Join(rootDirectory, "Shortest Video on Youtube Part 4.mp3"),
				filepath.Join(rootDirectory, "Shortest Video on Youtube Part 5.mp3"),
				filepath.Join(rootDirectory, "Shortest Video on Youtube Part 6.mp3"),
				filepath.Join(rootDirectory, "Shortest Video on Youtube Part 7.mp3"),
				filepath.Join(rootDirectory, "Shortest Video on Youtube Part 8.mp3"),
			},
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

func TestYoutubeAudioDownloader_IsVideoAvailable_Negative_Test(t *testing.T) {
	checkPrerequisites(t)
	downloader := NewYoutubeAudioDownloader()

	isAvailable := downloader.IsVideoAvailable("https://www.youtube.com/watch?v=invalid_url")
	if isAvailable {
		t.Errorf("Video is reported to available but should not be accessible")
	}
}

func TestYoutubeAudioDownloader_IsVideoAvailable(t *testing.T) {
	checkPrerequisites(t)
	downloader := NewYoutubeAudioDownloader()

	isAvailable := downloader.IsVideoAvailable(validYoutubeVideoUrl)
	if !isAvailable {
		t.Errorf("Video is reported to not available")
	}
}
