package filemanagement

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
)

// MoveToTarget moves sourcePath into a subdirectory of targetRootPath.
// The subdirectory name is taken from the immediate parent directory of sourcePath.
// e.g. sourcePath=/tmp/abc/channel/file.mp3, targetRootPath=/podcasts → /podcasts/channel/file.mp3
func MoveToTarget(sourcePath, targetRootPath string) (string, error) {
	directoryName := filepath.Base(filepath.Dir(sourcePath))
	targetSubDirectory := filepath.Join(targetRootPath, directoryName)
	if err := os.MkdirAll(targetSubDirectory, os.ModePerm); err != nil {
		return "", err
	}

	targetPath := filepath.Join(targetSubDirectory, filepath.Base(sourcePath))
	if err := MoveFile(sourcePath, targetPath); err != nil {
		return "", err
	}
	return targetPath, nil
}

func calculateFileHash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return make([]byte, 0), err
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("Error closing file", "err", err)
		}
	}()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return make([]byte, 0), err
	}
	return hasher.Sum(nil), err
}

func areFileEqual(leftFilePath string, rightFilePath string) (bool, error) {
	leftFileHash, err := calculateFileHash(leftFilePath)
	if err != nil {
		return false, err
	}
	slog.Info("creating hash of pre-existing file in target location")
	rightFileHash, err := calculateFileHash(rightFilePath)
	if err != nil {
		return false, err
	}
	return bytes.Equal(leftFileHash, rightFileHash), nil
}

func doesFileExist(filePath string) bool {
	_, err := os.Stat(filePath)
	return !errors.Is(err, os.ErrNotExist)
}

func MoveFile(sourcePath, targetPath string) (err error) {
	inputFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	inputClosed := false
	closeInput := func() {
		if !inputClosed {
			inputClosed = true
			if cerr := inputFile.Close(); cerr != nil {
				slog.Warn("Error closing input file", "err", cerr)
			}
		}
	}
	defer closeInput()

	fileName := filepath.Base(inputFile.Name())

	if doesFileExist(targetPath) {
		slog.Info("file already exists at target", "fileName", fileName)
		slog.Info("creating hash of original file")
		filesEqual, err := areFileEqual(sourcePath, targetPath)
		if err != nil {
			return err
		}
		if filesEqual {
			slog.Info("hash equal, deleting file at origin")
			// Must close inputFile before os.Remove on Windows
			closeInput()
			return os.Remove(sourcePath)
		}
		slog.Info("hash not equal, deleting file from target folder")
		if err = os.Remove(targetPath); err != nil {
			return err
		}
	}

	tempFileName := fmt.Sprintf("%s.part", targetPath)
	outputFile, err := os.Create(tempFileName)
	if err != nil {
		return err
	}
	defer func() {
		slog.Info("securely cleaning cache and closing file")
		if cerr := outputFile.Close(); cerr != nil {
			slog.Warn("could not close output file", "err", cerr)
			return
		}

		if err != nil {
			return
		}

		slog.Info("renaming temp file", "targetPath", targetPath)
		err = os.Rename(tempFileName, targetPath)
		if err != nil {
			slog.Error("could not rename file", "err", err)
			return
		}

		slog.Info("file moved successfully", "fileName", fileName)
		slog.Info("deleting file from source folder", "fileName", fileName)
		// Must close inputFile before os.Remove on Windows
		closeInput()
		if rerr := os.Remove(sourcePath); rerr != nil {
			slog.Warn("could not remove source file", "err", rerr)
			err = rerr
		}
	}()

	_, err = io.Copy(outputFile, inputFile)
	return err
}
