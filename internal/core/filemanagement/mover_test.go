package filemanagement

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func setupTestMoveEnvironment(t *testing.T) (rootDirectory string, leftDirectory string, rightDirectory string, fileName string) {
	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	if err != nil {
		t.Error("could not create folder")
	}

	leftDirectory, err = os.MkdirTemp(rootDirectory, "left")
	if err != nil {
		t.Error("could not create folder")
	}
	rightDirectory, err = os.MkdirTemp(rootDirectory, "right")
	if err != nil {
		t.Error("could not create folder")
	}
	file, err := os.CreateTemp(leftDirectory, "testFile")
	if err != nil {
		t.Error("could not create file")
	}
	fileName = filepath.Base(file.Name())
	if err := file.Close(); err != nil {
		t.Errorf("Error closing file: %v", err)
	}

	return rootDirectory, leftDirectory, rightDirectory, fileName
}

func TestMoveToTarget(t *testing.T) {
	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	if err != nil {
		t.Fatal("could not create root folder")
	}
	defer func() {
		if err := os.RemoveAll(rootDirectory); err != nil {
			t.Errorf("could not delete root directory: %v", err)
		}
	}()

	// source: rootDirectory/channel/file.mp3
	sourceDir := filepath.Join(rootDirectory, "channel")
	if err := os.Mkdir(sourceDir, os.ModePerm); err != nil {
		t.Fatal("could not create source directory")
	}
	sourceFile, err := os.CreateTemp(sourceDir, "*.mp3")
	if err != nil {
		t.Fatal("could not create source file")
	}
	sourcePath := sourceFile.Name()
	if err := sourceFile.Close(); err != nil {
		t.Fatalf("could not close source file: %v", err)
	}

	targetRoot, err := os.MkdirTemp(os.TempDir(), "targetDir")
	if err != nil {
		t.Fatal("could not create target root")
	}
	defer func() {
		if err := os.RemoveAll(targetRoot); err != nil {
			t.Errorf("could not delete target directory: %v", err)
		}
	}()

	result, err := MoveToTarget(sourcePath, targetRoot)
	if err != nil {
		t.Fatalf("MoveToTarget() error = %v", err)
	}

	expectedPath := filepath.Join(targetRoot, "channel", filepath.Base(sourcePath))
	if result != expectedPath {
		t.Errorf("MoveToTarget() = %q, want %q", result, expectedPath)
	}
	if _, err := os.Stat(result); errors.Is(err, os.ErrNotExist) {
		t.Errorf("file not found at target path %q", result)
	}
	if _, err := os.Stat(sourcePath); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("source file still exists at %q after move", sourcePath)
	}
}

func TestMoveFile(t *testing.T) {
	rootDirectory, leftDirectory, rightDirectory, fileName := setupTestMoveEnvironment(t)
	// clean-up
	defer func() {
		err := os.RemoveAll(rootDirectory)
		if err != nil {
			t.Errorf("could not delete file '%+v'", err)
		}
	}()
	target := filepath.Join(rightDirectory, fileName)
	origin := filepath.Join(leftDirectory, fileName)

	err := MoveFile(origin, target)
	if err != nil {
		t.Errorf("found error '%+v'", err)
	}
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		t.Errorf("file '%s' was not found", fileName)
	}
	if _, err := os.Stat(origin); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("file '%s' was removed", fileName)
	}
}

func TestMoveFileOnToExistingFile(t *testing.T) {
	rootDirectory, leftDirectory, rightDirectory, fileName := setupTestMoveEnvironment(t)
	// clean-up
	defer func() {
		err := os.RemoveAll(rootDirectory)
		if err != nil {
			t.Errorf("could not delete file '%+v'", err)
		}
	}()
	target := filepath.Join(rightDirectory, fileName)
	origin := filepath.Join(leftDirectory, fileName)

	file, err := os.Create(target)
	if err != nil {
		t.Error("could not create file")
	}
	if err := file.Close(); err != nil {
		t.Errorf("Error closing file: %v", err)
	}

	err = MoveFile(origin, target)
	if err != nil {
		t.Errorf("found error '%+v'", err)
	}
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		t.Errorf("file '%s' was not found", fileName)
	}
	if _, err := os.Stat(origin); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("file '%s' was removed", fileName)
	}
}

func TestMoveFileOnToCorruptExistingFile(t *testing.T) {
	rootDirectory, leftDirectory, rightDirectory, fileName := setupTestMoveEnvironment(t)
	// clean-up
	defer func() {
		err := os.RemoveAll(rootDirectory)
		if err != nil {
			t.Errorf("could not delete file '%+v'", err)
		}
	}()
	target := filepath.Join(rightDirectory, fileName)
	origin := filepath.Join(leftDirectory, fileName)

	file, err := os.Create(target)
	if err != nil {
		t.Error("could not create file")
	}
	_, err = file.WriteString("corrupt")
	if err != nil {
		t.Error("could not write to file")
	}
	if err := file.Close(); err != nil {
		t.Errorf("Error closing file: %v", err)
	}

	err = MoveFile(origin, target)
	if err != nil {
		t.Errorf("found error '%+v'", err)
	}
	if _, err := os.Stat(target); errors.Is(err, os.ErrNotExist) {
		t.Errorf("file '%s' was not found", fileName)
	}
	if _, err := os.Stat(origin); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("file '%s' was removed", fileName)
	}
}
