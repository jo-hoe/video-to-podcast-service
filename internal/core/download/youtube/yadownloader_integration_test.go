package youtube

import (
	"os"
	"path/filepath"
	"testing"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
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
	defer func() {
		if err := os.RemoveAll(rootDirectory); err != nil {
			t.Errorf("Error removing temp directory: %v", err)
		}
	}()
	if err != nil {
		t.Error("could not create folder")
	}

	// Create temp directory for download processing
	tempDir, err := os.MkdirTemp(os.TempDir(), "testTempDir")
	if err != nil {
		t.Fatalf("could not create temp folder: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Error removing temp directory: %v", err)
		}
	}()

	y := NewYoutubeAudioDownloader(nil, &config.Media{TempPath: tempDir})
	result, err := y.Download(validYoutubeVideoUrl, rootDirectory)
	if err != nil {
		t.Fatalf("YoutubeAudioDownloader.Download() error = %v", err)
	}
	if result == "" {
		t.Fatalf("YoutubeAudioDownloader.Download() returned empty result")
	}

	// check if metadata is set
	metadata, err := mp3joiner.GetFFmpegMetadataTag(result)
	if err != nil {
		t.Fatalf("YoutubeAudioDownloader.Download() error = %v", err)
	}
	expectedArtist := "jawed"
	if metadata["artist"] != expectedArtist {
		t.Errorf("YoutubeAudioDownloader.Download() artist = %v, want %v", metadata["artist"], expectedArtist)
	}
	if metadata[downloader.ThumbnailUrlTag] == "" {
		t.Errorf("YoutubeAudioDownloader.Download() thumbnail url tag was empty")
	}
	if metadata[downloader.PodcastDescriptionTag] == "" {
		t.Errorf("YoutubeAudioDownloader.Download() podcast description was empty")
	}
	if metadata[downloader.DateTag] == "" {
		t.Errorf("YoutubeAudioDownloader.Download() date tag was empty")
	}
	if metadata[downloader.VideoDownloadLink] != validYoutubeVideoUrl {
		t.Errorf("YoutubeAudioDownloader.Download() video url tag mismatch = %v", metadata[downloader.VideoDownloadLink])
	}

	// check if file is saved in correct location
	expectedFilename := "Me at the zoo_jNQXAC9IVRw.mp3"
	if result != filepath.Join(rootDirectory, expectedArtist, expectedFilename) {
		t.Errorf("YoutubeAudioDownloader.Download() path = %v, want %v", result, filepath.Join(rootDirectory, expectedArtist, expectedFilename))
	}

	chapters, err := mp3joiner.GetChapterMetadata(result)
	if err != nil {
		t.Fatalf("YoutubeAudioDownloader.Download() error = %v", err)
	}
	if len(chapters) < 1 {
		t.Error("YoutubeAudioDownloader.Download() no chapters have been found")
	}
}

func Test_YoutubeAudioDownloader_Download(t *testing.T) {
	checkPrerequisites(t)

	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	defer func() {
		if err := os.RemoveAll(rootDirectory); err != nil {
			t.Errorf("Error removing temp directory: %v", err)
		}
	}()
	if err != nil {
		t.Error("could not create folder")
	}

	// Create temp directory for download processing
	tempDir, err := os.MkdirTemp(os.TempDir(), "testTempDir")
	if err != nil {
		t.Fatalf("could not create temp folder: %v", err)
	}
	defer func() {
		if err := os.RemoveAll(tempDir); err != nil {
			t.Errorf("Error removing temp directory: %v", err)
		}
	}()

	y := NewYoutubeAudioDownloader(nil, &config.Media{TempPath: tempDir})

	// Single video download should return a single file path and file should exist
	singleResult, err := y.Download(validYoutubeVideoUrl, rootDirectory)
	if err != nil {
		t.Fatalf("YoutubeAudioDownloader.Download(single) error = %v", err)
	}
	if singleResult == "" {
		t.Fatalf("YoutubeAudioDownloader.Download(single) returned empty path")
	}
	if _, statErr := os.Stat(singleResult); statErr != nil {
		t.Fatalf("Downloaded file does not exist at path: %s, err: %v", singleResult, statErr)
	}

	// Playlist: list entries, download each entry individually, verify count and existence
	entries, err := y.ListIndividualVideoURLs(validYoutubePlaylistUrl)
	if err != nil {
		t.Fatalf("ListIndividualVideoURLs() error = %v", err)
	}
	if len(entries) == 0 {
		t.Fatalf("ListIndividualVideoURLs() returned no entries for valid playlist")
	}

	results := make([]string, 0, len(entries))
	for _, entry := range entries {
		p, err := y.Download(entry, filepath.Join(rootDirectory, "Cat"))
		if err != nil {
			t.Fatalf("Download(entry) error = %v", err)
		}
		if p == "" {
			t.Fatalf("Download(entry) returned empty path")
		}
		if _, statErr := os.Stat(p); statErr != nil {
			t.Fatalf("Downloaded file does not exist at path: %s, err: %v", p, statErr)
		}
		results = append(results, p)
	}

	// Expect at least multiple items; playlist contains multiple known entries (8 at the time of writing)
	if len(results) < 2 {
		t.Errorf("Expected multiple downloaded items for playlist, got %d", len(results))
	}
}

func TestYoutubeAudioDownloader_IsVideoAvailable_Negative_Test(t *testing.T) {
	checkPrerequisites(t)
	downloader := NewYoutubeAudioDownloader(nil, nil)

	isAvailable := downloader.IsVideoAvailable("https://www.youtube.com/watch?v=invalid_url")
	if isAvailable {
		t.Errorf("Video is reported to available but should not be accessible")
	}
}

func TestYoutubeAudioDownloader_IsVideoAvailable(t *testing.T) {
	checkPrerequisites(t)
	downloader := NewYoutubeAudioDownloader(nil, nil)

	isAvailable := downloader.IsVideoAvailable(validYoutubeVideoUrl)
	if !isAvailable {
		t.Errorf("Video is reported to not available")
	}
}
