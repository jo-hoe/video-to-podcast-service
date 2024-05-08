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
	file.Close()

	if err != nil {
		t.Errorf("could not close file %+v", err)
	}

	return rootDirectory, leftDirectory, rightDirectory, fileName
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
	file.Close()

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
	file.Close()

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