package database

import (
	"fmt"
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

func TestInsertReplacePodcastItem(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)
	item := getDemoPodcastItem()
	err := db.InsertReplacePodcastItem(item)
	if err != nil {
		t.Fatalf("failed to create podcast item: %v", err)
	}
	fetched, err := getPodcastItemByID(t, db, item.ID)
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

	err := db.InsertReplacePodcastItem(item)
	if err != nil {
		t.Fatalf("failed to create podcast item: %v", err)
	}
	err = db.InsertReplacePodcastItem(updated)
	if err != nil {
		t.Fatalf("failed to update podcast item: %v", err)
	}

	foundItem, err := getPodcastItemByID(t, db, item.ID)

	if err != nil {
		t.Fatalf("failed to fetch podcast item: %v", err)
	}
	if foundItem == nil {
		t.Fatal("podcast item not found after update")
	}
	if foundItem.DurationInMilliseconds != updated.DurationInMilliseconds {
		t.Errorf("expected duration %d, got %d", updated.DurationInMilliseconds, foundItem.DurationInMilliseconds)
	}
}

func getPodcastItemByID(t *testing.T, db *SQLiteDatabase, id string) (*PodcastItem, error) {
	allItems, err := db.GetAllPodcastItems()
	if err != nil {
		return nil, err
	}
	foundItem := &PodcastItem{}
	for _, fetched := range allItems {
		if fetched.ID == id {
			foundItem = fetched
			break
		}
	}
	if foundItem.ID == "" {
		return nil, fmt.Errorf("podcast item with ID %s not found", id)
	}
	return foundItem, nil
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
