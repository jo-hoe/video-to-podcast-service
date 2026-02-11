package filemanagement

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetAudioFiles_FindsMP3Files(t *testing.T) {
	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "..", "test_assets"))
	if err != nil {
		t.Fatalf("failed to resolve test assets path: %v", err)
	}

	gotResult, err := GetAudioFiles(testFilePath)
	if err != nil {
		t.Fatalf("GetAudioFiles() unexpected error: %v", err)
	}
	wantResult := []string{
		filepath.Join(testFilePath, "audio11.mp3"),
		filepath.Join(testFilePath, "audio12.mp3"),
		filepath.Join(testFilePath, "audio21.mp3"),
	}
	if !reflect.DeepEqual(gotResult, wantResult) {
		t.Fatalf("GetAudioFiles() = %v, want %v", gotResult, wantResult)
	}
}

func TestGetAudioFiles_NonExistingDirectoryReturnsError(t *testing.T) {
	nonExisting := "test/non-existing-directory"
	gotResult, err := GetAudioFiles(nonExisting)
	if err == nil {
		t.Fatalf("GetAudioFiles() expected error for non-existing directory, got nil")
	}
	if gotResult == nil {
		t.Fatalf("GetAudioFiles() expected empty slice, got nil")
	}
	if len(gotResult) != 0 {
		t.Fatalf("GetAudioFiles() expected 0 results, got %d", len(gotResult))
	}
}