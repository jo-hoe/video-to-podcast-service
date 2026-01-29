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
	defer func() {
		if err := inputFile.Close(); err != nil {
			slog.Warn("Error closing input file", "err", err)
		}
	}()
	fileName := filepath.Base(inputFile.Name())

	if doesFileExist(targetPath) {
		slog.Info("file already exists at target", "fileName", fileName)
		slog.Info("creating hash of original file")
		// check if this is the same file
		filesEqual, err := areFileEqual(sourcePath, targetPath)
		if err != nil {
			return err
		}
		if filesEqual {
			slog.Info("hash equal, deleting file at origin")
			// same file already exists and can be removed from source
			err = inputFile.Close()
			if err != nil {
				return err
			}
			err = os.Remove(sourcePath)
			if err != nil {
				return err
			}
			// stop process
			return nil
		} else {
			slog.Info("hash not equal, deleting file from target folder")
			// remove destination file and continue
			err = os.Remove(targetPath)
			if err != nil {
				return err
			}
		}
	}

	tempFileName := fmt.Sprintf("%s.part", targetPath)
	outputFile, err := os.Create(tempFileName)
	if err != nil {
		if err := inputFile.Close(); err != nil {
			slog.Warn("Error closing input file", "err", err)
		}
		return err
	}
	defer func() {
		slog.Info("securely cleaning cache and closing file")
		fileClosingError := outputFile.Close()
		if fileClosingError != nil {
			slog.Warn("could not close file")
			return
		}

		// check if copying was successful
		if err != nil {
			return
		}

		slog.Info("renaming temp file", "targetPath", targetPath)
		// rename file to actual file name
		err = os.Rename(tempFileName, targetPath)
		if err != nil {
			slog.Error("could not rename file", "err", err)
			return
		}

		// currently double-check of file hash
		// target/source file has been omitted

		slog.Info("file moved successfully", "fileName", fileName)
		// The copy was successful, so now delete the original file
		slog.Info("deleting file from source folder", "fileName", fileName)
		err = os.Remove(sourcePath)
		if err != nil {
			slog.Warn("could not close file", "err", err)
		}

	}()

	// actual file copy
	_, err = io.Copy(outputFile, inputFile)
	if err := inputFile.Close(); err != nil {
		slog.Warn("Error closing input file", "err", err)
	}
	if err != nil {
		return err
	}

	return nil
}
