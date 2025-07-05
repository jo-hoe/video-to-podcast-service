package database

import (
	"testing"
	"time"
)

const (
	testDBFile   = "test_podcast_items.db"
	testItemID   = "test-id"
	testTitle    = "Test Title"
	testDesc     = "Test Description"
	testAuthor   = "Test Author"
	testThumb    = "test_thumbnail.jpg"
	testVideoURL = "http://example.com/video"
	testAudio    = "audio.mp3"
	duration     = 120000 // 2 minutes in milliseconds
)

func TestCreatePodcastItem(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	item := getDemoPodcastItem()
	err := db.CreatePodcastItem(item)
	if err != nil {
		t.Fatalf("failed to create podcast item: %v", err)
	}
	fetched, err := db.GetPodcastItemByID(item.ID)
	if err != nil {
		t.Fatalf("failed to fetch podcast item: %v", err)
	}
	if fetched == nil {
		t.Fatal("podcast item not found after creation")
	}
	if fetched.Title != item.Title {
		t.Errorf("expected title %q, got %q", item.Title, fetched.Title)
	}
}

func TestCreatePodcastItemTwice(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	item := getDemoPodcastItem()
	updated := &PodcastItem{
		ID:                     item.ID,
		Title:                  item.Title,
		Description:            item.Description,
		Author:                 item.Author,
		Thumbnail:              item.Thumbnail,
		DurationInMilliseconds: 150000, // Updated duration
		VideoURL:               item.VideoURL,
		AudioFilePath:          item.AudioFilePath,
		CreatedAt:              item.CreatedAt,
		UpdatedAt:              time.Now(), // Updated time
	}

	err := db.CreatePodcastItem(item)
	if err != nil {
		t.Fatalf("failed to create podcast item: %v", err)
	}
	err = db.CreatePodcastItem(updated)
	if err != nil {
		t.Fatalf("failed to update podcast item: %v", err)
	}
	fetched, err := db.GetPodcastItemByID(testItemID)
	if err != nil {
		t.Fatalf("failed to fetch podcast item: %v", err)
	}
	if fetched == nil {
		t.Fatal("podcast item not found after update")
	}
	if fetched.DurationInMilliseconds != updated.DurationInMilliseconds {
		t.Errorf("expected duration %d, got %d", updated.DurationInMilliseconds, fetched.DurationInMilliseconds)
	}
}

func getDemoPodcastItem() *PodcastItem {
	return &PodcastItem{
		ID:                     testItemID,
		Title:                  testTitle,
		Description:            testDesc,
		Author:                 testAuthor,
		Thumbnail:              testThumb,
		DurationInMilliseconds: duration,
		VideoURL:               testVideoURL,
		AudioFilePath:          testAudio,
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
}

func setupTestDB(t *testing.T) *SQLiteDatabase {
	db := NewSQLiteDatabase(testDBFile)
	_, err := db.CreateDatabase()
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	return db
}

func cleanupTestDB(t *testing.T, db *SQLiteDatabase) {
	err := db.DropDatabase()
	if err != nil {
		t.Fatalf("failed to drop database: %v", err)
	}
}
