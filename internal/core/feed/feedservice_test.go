package feed

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
)

func TestCreateFeed(t *testing.T) {
	defaultAuthor := "John Doe"
	type fields struct {
		feedBasePort      string
		feedItemPath      string
		feedAudioFilePath string
		feedAuthor        string
		feedHost          string
		coreService       *core.CoreService
	}
	tests := []struct {
		name   string
		fields fields
		want   *feeds.Feed
	}{
		{
			name: "create feed test",
			fields: fields{
				feedBasePort:      "443",
				feedItemPath:      "v1/feeds",
				feedAuthor:        defaultAuthor,
				feedHost:          "localhost",
				feedAudioFilePath: filepath.Join("c", "testDir", "audio.mp3"),
				coreService:       core.NewCoreService(&database.MockDatabase{}, filepath.Join("c")),
			},
			want: &feeds.Feed{
				Title:       defaultAuthor,
				Link:        &feeds.Link{Href: "http://localhost/v1/feeds/testDir/rss.xml"},
				Description: fmt.Sprintf("%s %s", defaultDescription, defaultAuthor),
				Author:      &feeds.Author{Name: defaultAuthor},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &FeedService{
				coreservice:  tt.fields.coreService,
				feedBasePort: tt.fields.feedBasePort,
				feedItemPath: tt.fields.feedItemPath,
				feedConfig:   &config.FeedConfig{Mode: "per_directory"},
			}
			if got := fp.createFeed(tt.fields.feedHost, tt.fields.feedAuthor, tt.fields.feedAudioFilePath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFeedService(t *testing.T) {
	type args struct {
		coreService  *core.CoreService
		feedBasePort string
		feedItemPath string
		feedConfig   *config.FeedConfig
	}
	sharedCore := core.NewCoreService(&database.MockDatabase{}, "testDir")
	tests := []struct {
		name string
		args args
		want *FeedService
	}{
		{
			name: "init test with per_directory config",
			args: args{
				coreService:  sharedCore,
				feedBasePort: "8080",
				feedItemPath: "v1/path",
				feedConfig:   &config.FeedConfig{Mode: "per_directory"},
			},
			want: &FeedService{
				coreservice:  sharedCore,
				feedBasePort: "8080",
				feedItemPath: "v1/path",
				feedConfig:   &config.FeedConfig{Mode: "per_directory"},
			},
		},
		{
			name: "init test with unified config",
			args: args{
				coreService:  sharedCore,
				feedBasePort: "8080",
				feedItemPath: "v1/path",
				feedConfig:   &config.FeedConfig{Mode: "unified"},
			},
			want: &FeedService{
				coreservice:  sharedCore,
				feedBasePort: "8080",
				feedItemPath: "v1/path",
				feedConfig:   &config.FeedConfig{Mode: "unified"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFeedService(tt.args.coreService, tt.args.feedBasePort, tt.args.feedItemPath, tt.args.feedConfig)
			if got.feedBasePort != tt.want.feedBasePort || got.feedItemPath != tt.want.feedItemPath {
				t.Errorf("NewFeedService() = %v, want %v", got, tt.want)
			}
			// Compare coreService by pointer address (since DeepEqual will fail on different instances)
			if got.coreservice != tt.want.coreservice {
				t.Errorf("NewFeedService() coreService pointer mismatch")
			}
			// Compare feedConfig
			if got.feedConfig.Mode != tt.want.feedConfig.Mode {
				t.Errorf("NewFeedService() feedConfig.Mode = %v, want %v", got.feedConfig.Mode, tt.want.feedConfig.Mode)
			}
		})
	}
}

func TestFeedService_GetFeeds_PerDirectoryMode(t *testing.T) {
	mockDB := &database.MockDatabase{}
	coreService := core.NewCoreService(mockDB, "testDir")
	feedConfig := &config.FeedConfig{Mode: "per_directory"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	// Test that GetFeeds routes to per-directory mode
	feeds, err := feedService.GetFeeds("localhost")
	if err != nil {
		t.Errorf("GetFeeds() error = %v", err)
		return
	}

	// Should return multiple feeds (one per directory) or empty slice if no items
	if feeds == nil {
		t.Errorf("GetFeeds() returned nil feeds")
	}
}

func TestFeedService_GetFeeds_UnifiedMode(t *testing.T) {
	mockDB := &database.MockDatabase{}
	coreService := core.NewCoreService(mockDB, "testDir")
	feedConfig := &config.FeedConfig{Mode: "unified"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	// Test that GetFeeds routes to unified mode
	feeds, err := feedService.GetFeeds("localhost")
	if err != nil {
		t.Errorf("GetFeeds() error = %v", err)
		return
	}

	// Should return exactly one feed in unified mode
	if len(feeds) != 1 {
		t.Errorf("GetFeeds() in unified mode returned %d feeds, want 1", len(feeds))
	}

	if len(feeds) > 0 {
		if feeds[0].Title != "All Podcast Items" {
			t.Errorf("GetFeeds() unified feed title = %v, want 'All Podcast Items'", feeds[0].Title)
		}
		if feeds[0].Description != "Unified podcast feed containing all items" {
			t.Errorf("GetFeeds() unified feed description = %v, want 'Unified podcast feed containing all items'", feeds[0].Description)
		}
	}
}

func TestFeedService_GetUnifiedFeed(t *testing.T) {
	mockDB := &database.MockDatabase{}
	coreService := core.NewCoreService(mockDB, "testDir")
	feedConfig := &config.FeedConfig{Mode: "unified"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	feed, err := feedService.getUnifiedFeed("localhost")
	if err != nil {
		t.Errorf("getUnifiedFeed() error = %v", err)
		return
	}

	if feed.Title != "All Podcast Items" {
		t.Errorf("getUnifiedFeed() title = %v, want 'All Podcast Items'", feed.Title)
	}

	if feed.Description != "Unified podcast feed containing all items" {
		t.Errorf("getUnifiedFeed() description = %v, want 'Unified podcast feed containing all items'", feed.Description)
	}

	if feed.Author.Name != "Video to Podcast Service" {
		t.Errorf("getUnifiedFeed() author = %v, want 'Video to Podcast Service'", feed.Author.Name)
	}

	expectedLink := "http://localhost:8080/v1/feeds/all/rss.xml"
	if feed.Link.Href != expectedLink {
		t.Errorf("getUnifiedFeed() link = %v, want %v", feed.Link.Href, expectedLink)
	}
}

func TestFeedService_GetUnifiedFeed_WithMultipleItems(t *testing.T) {
	// Create temporary directory and files for testing
	tempDir := t.TempDir()
	channel1Dir := filepath.Join(tempDir, "channel1")
	channel2Dir := filepath.Join(tempDir, "channel2")

	if err := os.MkdirAll(channel1Dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(channel2Dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	episode1Path := filepath.Join(channel1Dir, "episode1.mp3")
	episode2Path := filepath.Join(channel2Dir, "episode1.mp3")
	episode3Path := filepath.Join(channel1Dir, "episode2.mp3")

	// Create empty test files
	if err := os.WriteFile(episode1Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(episode2Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(episode3Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockDB := database.NewMockDatabase()

	// Create items with specific creation times to test sorting
	baseTime := time.Now()
	item1 := &database.PodcastItem{
		ID:            "test-id-1",
		Title:         "Channel 1 Episode 1",
		Description:   "Test Description 1",
		Author:        "Author 1",
		Thumbnail:     "http://example.com/thumb1.jpg",
		AudioFilePath: episode1Path,
		CreatedAt:     baseTime.Add(-2 * time.Hour), // Oldest
	}
	item2 := &database.PodcastItem{
		ID:            "test-id-2",
		Title:         "Channel 2 Episode 1",
		Description:   "Test Description 2",
		Author:        "Author 2",
		Thumbnail:     "http://example.com/thumb2.jpg",
		AudioFilePath: episode2Path,
		CreatedAt:     baseTime, // Newest
	}
	item3 := &database.PodcastItem{
		ID:            "test-id-3",
		Title:         "Channel 1 Episode 2",
		Description:   "Test Description 3",
		Author:        "Author 1",
		Thumbnail:     "http://example.com/thumb3.jpg",
		AudioFilePath: episode3Path,
		CreatedAt:     baseTime.Add(-1 * time.Hour), // Middle
	}

	mockDB.Items["test-id-1"] = item1
	mockDB.Items["test-id-2"] = item2
	mockDB.Items["test-id-3"] = item3

	coreService := core.NewCoreService(mockDB, tempDir)
	feedConfig := &config.FeedConfig{Mode: "unified"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	feed, err := feedService.getUnifiedFeed("localhost")
	if err != nil {
		t.Errorf("getUnifiedFeed() error = %v", err)
		return
	}

	// Verify unified feed properties
	if feed.Title != "All Podcast Items" {
		t.Errorf("getUnifiedFeed() title = %v, want 'All Podcast Items'", feed.Title)
	}

	if feed.Description != "Unified podcast feed containing all items" {
		t.Errorf("getUnifiedFeed() description = %v, want 'Unified podcast feed containing all items'", feed.Description)
	}

	if feed.Author.Name != "Video to Podcast Service" {
		t.Errorf("getUnifiedFeed() author = %v, want 'Video to Podcast Service'", feed.Author.Name)
	}

	expectedLink := "http://localhost:8080/v1/feeds/all/rss.xml"
	if feed.Link.Href != expectedLink {
		t.Errorf("getUnifiedFeed() link = %v, want %v", feed.Link.Href, expectedLink)
	}

	// Verify all items are included
	if len(feed.Items) != 3 {
		t.Errorf("getUnifiedFeed() has %d items, want 3", len(feed.Items))
		return
	}

	// Verify items are sorted by creation date (newest first)
	if len(feed.Items) >= 3 {
		// First item should be the newest (item2)
		if feed.Items[0].Id != "test-id-2" {
			t.Errorf("getUnifiedFeed() first item ID = %v, want 'test-id-2' (newest)", feed.Items[0].Id)
		}
		// Second item should be the middle one (item3)
		if feed.Items[1].Id != "test-id-3" {
			t.Errorf("getUnifiedFeed() second item ID = %v, want 'test-id-3' (middle)", feed.Items[1].Id)
		}
		// Third item should be the oldest (item1)
		if feed.Items[2].Id != "test-id-1" {
			t.Errorf("getUnifiedFeed() third item ID = %v, want 'test-id-1' (oldest)", feed.Items[2].Id)
		}
	}

	// Verify feed image is set from first item with thumbnail
	if feed.Image == nil {
		t.Errorf("getUnifiedFeed() feed image is nil, expected to be set")
	} else {
		expectedThumbnails := []string{"http://example.com/thumb1.jpg", "http://example.com/thumb2.jpg", "http://example.com/thumb3.jpg"}
		found := false
		for _, expectedThumb := range expectedThumbnails {
			if feed.Image.Url == expectedThumb {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getUnifiedFeed() feed image URL = %v, want one of %v", feed.Image.Url, expectedThumbnails)
		}
	}

	// Verify items from different directories are included
	foundChannel1Items := 0
	foundChannel2Items := 0
	for _, item := range feed.Items {
		switch item.Title {
		case "Channel 1 Episode 1", "Channel 1 Episode 2":
			foundChannel1Items++
		case "Channel 2 Episode 1":
			foundChannel2Items++
		}
	}

	if foundChannel1Items != 2 {
		t.Errorf("getUnifiedFeed() found %d channel1 items, want 2", foundChannel1Items)
	}
	if foundChannel2Items != 1 {
		t.Errorf("getUnifiedFeed() found %d channel2 items, want 1", foundChannel2Items)
	}
}

func TestFeedService_GetUnifiedFeed_EmptyDatabase(t *testing.T) {
	mockDB := &database.MockDatabase{}
	coreService := core.NewCoreService(mockDB, "testDir")
	feedConfig := &config.FeedConfig{Mode: "unified"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	feed, err := feedService.getUnifiedFeed("localhost")
	if err != nil {
		t.Errorf("getUnifiedFeed() error = %v", err)
		return
	}

	// Verify unified feed properties even with empty database
	if feed.Title != "All Podcast Items" {
		t.Errorf("getUnifiedFeed() title = %v, want 'All Podcast Items'", feed.Title)
	}

	if feed.Description != "Unified podcast feed containing all items" {
		t.Errorf("getUnifiedFeed() description = %v, want 'Unified podcast feed containing all items'", feed.Description)
	}

	if feed.Author.Name != "Video to Podcast Service" {
		t.Errorf("getUnifiedFeed() author = %v, want 'Video to Podcast Service'", feed.Author.Name)
	}

	// Should have no items
	if len(feed.Items) != 0 {
		t.Errorf("getUnifiedFeed() with empty database has %d items, want 0", len(feed.Items))
	}

	// Should have no image
	if feed.Image != nil {
		t.Errorf("getUnifiedFeed() with empty database has image, want nil")
	}
}

func TestFeedService_GetUnifiedFeed_SortingOrder(t *testing.T) {
	// Create temporary directory and files for testing
	tempDir := t.TempDir()
	channelDir := filepath.Join(tempDir, "channel")

	if err := os.MkdirAll(channelDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	episode1Path := filepath.Join(channelDir, "episode1.mp3")
	episode2Path := filepath.Join(channelDir, "episode2.mp3")
	episode3Path := filepath.Join(channelDir, "episode3.mp3")

	// Create empty test files
	for _, path := range []string{episode1Path, episode2Path, episode3Path} {
		if err := os.WriteFile(path, []byte("test audio data"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	mockDB := database.NewMockDatabase()

	// Create items with specific creation times to test sorting
	baseTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)

	// Add items in non-chronological order to test sorting
	item1 := &database.PodcastItem{
		ID:            "test-id-1",
		Title:         "Episode 1",
		Description:   "First episode",
		Author:        "Test Author",
		AudioFilePath: episode1Path,
		CreatedAt:     baseTime.Add(1 * time.Hour), // Middle time
	}
	item2 := &database.PodcastItem{
		ID:            "test-id-2",
		Title:         "Episode 2",
		Description:   "Second episode",
		Author:        "Test Author",
		AudioFilePath: episode2Path,
		CreatedAt:     baseTime.Add(3 * time.Hour), // Latest time
	}
	item3 := &database.PodcastItem{
		ID:            "test-id-3",
		Title:         "Episode 3",
		Description:   "Third episode",
		Author:        "Test Author",
		AudioFilePath: episode3Path,
		CreatedAt:     baseTime, // Earliest time
	}

	mockDB.Items["test-id-1"] = item1
	mockDB.Items["test-id-2"] = item2
	mockDB.Items["test-id-3"] = item3

	coreService := core.NewCoreService(mockDB, tempDir)
	feedConfig := &config.FeedConfig{Mode: "unified"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	feed, err := feedService.getUnifiedFeed("localhost")
	if err != nil {
		t.Errorf("getUnifiedFeed() error = %v", err)
		return
	}

	// Verify items are sorted by creation date (newest first)
	if len(feed.Items) != 3 {
		t.Errorf("getUnifiedFeed() has %d items, want 3", len(feed.Items))
		return
	}

	// Check that items are in correct order (newest to oldest)
	expectedOrder := []string{"test-id-2", "test-id-1", "test-id-3"}
	for i, expectedID := range expectedOrder {
		if feed.Items[i].Id != expectedID {
			t.Errorf("getUnifiedFeed() item %d ID = %v, want %v", i, feed.Items[i].Id, expectedID)
		}
	}

	// Verify the actual creation times are in descending order
	for i := 0; i < len(feed.Items)-1; i++ {
		if feed.Items[i].Created.Before(feed.Items[i+1].Created) {
			t.Errorf("getUnifiedFeed() items not sorted correctly: item %d (%v) is older than item %d (%v)",
				i, feed.Items[i].Created, i+1, feed.Items[i+1].Created)
		}
	}
}

func TestFeedService_GetPerDirectoryFeeds_EmptyDatabase(t *testing.T) {
	mockDB := &database.MockDatabase{}
	coreService := core.NewCoreService(mockDB, "testDir")
	feedConfig := &config.FeedConfig{Mode: "per_directory"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	feeds, err := feedService.getPerDirectoryFeeds("localhost")
	if err != nil {
		t.Errorf("getPerDirectoryFeeds() error = %v", err)
		return
	}

	if len(feeds) != 0 {
		t.Errorf("getPerDirectoryFeeds() with empty database returned %d feeds, want 0", len(feeds))
	}
}

func TestFeedService_GetPerDirectoryFeeds_SingleDirectory(t *testing.T) {
	// Create temporary directory and files for testing
	tempDir := t.TempDir()
	channel1Dir := filepath.Join(tempDir, "channel1")
	if err := os.MkdirAll(channel1Dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	episode1Path := filepath.Join(channel1Dir, "episode1.mp3")
	episode2Path := filepath.Join(channel1Dir, "episode2.mp3")

	// Create empty test files
	if err := os.WriteFile(episode1Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(episode2Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockDB := database.NewMockDatabase()

	// Add test podcast items from the same directory
	item1 := &database.PodcastItem{
		ID:            "test-id-1",
		Title:         "Test Episode 1",
		Description:   "Test Description 1",
		Author:        "Test Author",
		Thumbnail:     "http://example.com/thumb1.jpg",
		AudioFilePath: episode1Path,
		CreatedAt:     time.Now(),
	}
	item2 := &database.PodcastItem{
		ID:            "test-id-2",
		Title:         "Test Episode 2",
		Description:   "Test Description 2",
		Author:        "Test Author",
		Thumbnail:     "http://example.com/thumb2.jpg",
		AudioFilePath: episode2Path,
		CreatedAt:     time.Now(),
	}

	mockDB.Items["test-id-1"] = item1
	mockDB.Items["test-id-2"] = item2

	coreService := core.NewCoreService(mockDB, tempDir)
	feedConfig := &config.FeedConfig{Mode: "per_directory"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	feeds, err := feedService.getPerDirectoryFeeds("localhost")
	if err != nil {
		t.Errorf("getPerDirectoryFeeds() error = %v", err)
		return
	}

	// Should have exactly one feed for the single directory
	if len(feeds) != 1 {
		t.Errorf("getPerDirectoryFeeds() returned %d feeds, want 1", len(feeds))
		return
	}

	feed := feeds[0]
	if feed.Title != "channel1" {
		t.Errorf("getPerDirectoryFeeds() feed title = %v, want 'channel1'", feed.Title)
	}

	if feed.Author.Name != "channel1" {
		t.Errorf("getPerDirectoryFeeds() feed author = %v, want 'channel1'", feed.Author.Name)
	}

	if len(feed.Items) != 2 {
		t.Errorf("getPerDirectoryFeeds() feed has %d items, want 2", len(feed.Items))
	}

	// Verify feed image is set from one of the items with thumbnail
	if feed.Image == nil {
		t.Errorf("getPerDirectoryFeeds() feed image is nil, expected to be set")
	} else {
		expectedThumbnails := []string{"http://example.com/thumb1.jpg", "http://example.com/thumb2.jpg"}
		found := false
		for _, expectedThumb := range expectedThumbnails {
			if feed.Image.Url == expectedThumb {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("getPerDirectoryFeeds() feed image URL = %v, want one of %v", feed.Image.Url, expectedThumbnails)
		}
	}
}

func TestFeedService_GetPerDirectoryFeeds_MultipleDirectories(t *testing.T) {
	// Create temporary directory and files for testing
	tempDir := t.TempDir()
	channel1Dir := filepath.Join(tempDir, "channel1")
	channel2Dir := filepath.Join(tempDir, "channel2")

	if err := os.MkdirAll(channel1Dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	if err := os.MkdirAll(channel2Dir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	episode1Path := filepath.Join(channel1Dir, "episode1.mp3")
	episode2Path := filepath.Join(channel2Dir, "episode1.mp3")
	episode3Path := filepath.Join(channel1Dir, "episode2.mp3")

	// Create empty test files
	if err := os.WriteFile(episode1Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(episode2Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(episode3Path, []byte("test audio data"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	mockDB := database.NewMockDatabase()

	// Add test podcast items from different directories
	item1 := &database.PodcastItem{
		ID:            "test-id-1",
		Title:         "Channel 1 Episode 1",
		Description:   "Test Description 1",
		Author:        "Author 1",
		Thumbnail:     "http://example.com/thumb1.jpg",
		AudioFilePath: episode1Path,
		CreatedAt:     time.Now(),
	}
	item2 := &database.PodcastItem{
		ID:            "test-id-2",
		Title:         "Channel 2 Episode 1",
		Description:   "Test Description 2",
		Author:        "Author 2",
		Thumbnail:     "http://example.com/thumb2.jpg",
		AudioFilePath: episode2Path,
		CreatedAt:     time.Now(),
	}
	item3 := &database.PodcastItem{
		ID:            "test-id-3",
		Title:         "Channel 1 Episode 2",
		Description:   "Test Description 3",
		Author:        "Author 1",
		Thumbnail:     "http://example.com/thumb3.jpg",
		AudioFilePath: episode3Path,
		CreatedAt:     time.Now(),
	}

	mockDB.Items["test-id-1"] = item1
	mockDB.Items["test-id-2"] = item2
	mockDB.Items["test-id-3"] = item3

	coreService := core.NewCoreService(mockDB, tempDir)
	feedConfig := &config.FeedConfig{Mode: "per_directory"}

	feedService := NewFeedService(coreService, ":8080", "/v1/feeds", feedConfig)

	feedList, err := feedService.getPerDirectoryFeeds("localhost")
	if err != nil {
		t.Errorf("getPerDirectoryFeeds() error = %v", err)
		return
	}

	// Should have exactly two feeds for the two directories
	if len(feedList) != 2 {
		t.Errorf("getPerDirectoryFeeds() returned %d feeds, want 2", len(feedList))
		return
	}

	// Find feeds by title (directory name)
	var channel1Feed, channel2Feed *feeds.Feed
	for _, feed := range feedList {
		switch feed.Title {
		case "channel1":
			channel1Feed = feed
		case "channel2":
			channel2Feed = feed
		}
	}

	if channel1Feed == nil {
		t.Errorf("getPerDirectoryFeeds() missing channel1 feed")
		return
	}
	if channel2Feed == nil {
		t.Errorf("getPerDirectoryFeeds() missing channel2 feed")
		return
	}

	// Verify channel1 feed has 2 items
	if len(channel1Feed.Items) != 2 {
		t.Errorf("getPerDirectoryFeeds() channel1 feed has %d items, want 2", len(channel1Feed.Items))
	}

	// Verify channel2 feed has 1 item
	if len(channel2Feed.Items) != 1 {
		t.Errorf("getPerDirectoryFeeds() channel2 feed has %d items, want 1", len(channel2Feed.Items))
	}

	// Verify feed properties
	if channel1Feed.Author.Name != "channel1" {
		t.Errorf("getPerDirectoryFeeds() channel1 feed author = %v, want 'channel1'", channel1Feed.Author.Name)
	}
	if channel2Feed.Author.Name != "channel2" {
		t.Errorf("getPerDirectoryFeeds() channel2 feed author = %v, want 'channel2'", channel2Feed.Author.Name)
	}

	// Verify feed descriptions
	expectedDesc1 := fmt.Sprintf("%s %s", defaultDescription, "channel1")
	expectedDesc2 := fmt.Sprintf("%s %s", defaultDescription, "channel2")
	if channel1Feed.Description != expectedDesc1 {
		t.Errorf("getPerDirectoryFeeds() channel1 feed description = %v, want %v", channel1Feed.Description, expectedDesc1)
	}
	if channel2Feed.Description != expectedDesc2 {
		t.Errorf("getPerDirectoryFeeds() channel2 feed description = %v, want %v", channel2Feed.Description, expectedDesc2)
	}
}
