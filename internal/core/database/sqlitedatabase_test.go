package database

import (
	"os"
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
)

func TestCreatePodcastItem(t *testing.T) {
	defer cleanupTestDB(t)
	db := setupTestDB(t)
	defer closeDB(t, db)
	item := getDemoPodcastItem(2)
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
	defer cleanupTestDB(t)
	db := setupTestDB(t)
	defer closeDB(t, db)
	item := getDemoPodcastItem(1)
	updated := getDemoPodcastItem(2)

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

func getDemoPodcastItem(duration int64) *PodcastItem {
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
	}
}

func setupTestDB(t *testing.T) *SQLiteDatabase {
	db := &SQLiteDatabase{}
	_, err := db.CreateDatabase(testDBFile)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	return db
}

func cleanupTestDB(t *testing.T) {
	err := os.Remove(testDBFile)
	if err != nil && !os.IsNotExist(err) {
		t.Fatalf("failed to remove test database file: %v", err)
	}
}

func closeDB(t *testing.T, db *SQLiteDatabase) {
	err := db.Close()
	if err != nil {
		t.Fatalf("failed to close database: %v", err)
	}
}
