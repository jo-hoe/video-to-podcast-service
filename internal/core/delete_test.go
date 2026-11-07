package core

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
)

func TestDeletePodcastItem_UnknownID_ReturnsError(t *testing.T) {
	db := database.NewMockDatabase()
	db.GetPodcastItemByIDFunc = func(_ string) (*database.PodcastItem, error) {
		return nil, os.ErrNotExist
	}
	cs := NewCoreService(db, t.TempDir(), nil, nil, nil)

	if err := cs.DeletePodcastItem("missing"); err == nil {
		t.Fatal("expected error when item lookup fails")
	}
}

func TestDeletePodcastItem_RemovesFileAndEmptyDir(t *testing.T) {
	root := t.TempDir()
	feedDir := filepath.Join(root, "channel")
	if err := os.MkdirAll(feedDir, 0755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	audioPath := filepath.Join(feedDir, "episode.mp3")
	if err := os.WriteFile(audioPath, []byte("audio"), 0644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	db := database.NewMockDatabase()
	db.Items["item-1"] = &database.PodcastItem{ID: "item-1", AudioFilePath: audioPath}
	cs := NewCoreService(db, root, nil, nil, nil)

	if err := cs.DeletePodcastItem("item-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(audioPath); !os.IsNotExist(err) {
		t.Error("expected audio file to be removed")
	}
	if _, err := os.Stat(feedDir); !os.IsNotExist(err) {
		t.Error("expected empty feed directory to be removed")
	}
	if _, err := db.GetPodcastItemByID("item-1"); err != nil || db.Items["item-1"] != nil {
		t.Error("expected db entry to be removed")
	}
}

func TestDeletePodcastItem_MissingAudioFile_StillDeletesEntry(t *testing.T) {
	root := t.TempDir()
	db := database.NewMockDatabase()
	db.Items["item-1"] = &database.PodcastItem{
		ID:            "item-1",
		AudioFilePath: filepath.Join(root, "channel", "gone.mp3"),
	}
	cs := NewCoreService(db, root, nil, nil, nil)

	if err := cs.DeletePodcastItem("item-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if db.Items["item-1"] != nil {
		t.Error("expected db entry to be removed even when file is missing")
	}
}
