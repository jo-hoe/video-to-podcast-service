package filemanagement

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

func calculateFileHash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return make([]byte, 0), err
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
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
	log.Print("creating hash of pre-existing file in target location")
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
	fileName := filepath.Base(inputFile.Name())

	// Close the input file immediately after opening
	if err := inputFile.Close(); err != nil {
		log.Printf("Error closing input file: %v", err)
	}

	if doesFileExist(targetPath) {
		log.Printf("file '%s' already exists at target", fileName)
		log.Print("creating hash of original file")
		// check if this is the same file
		filesEqual, err := areFileEqual(sourcePath, targetPath)
		if err != nil {
			return err
		}
		if filesEqual {
			log.Print("hash was equal, deleting file at origin")
			// same file already exists and can be removed from source
			if err := removeFile(sourcePath); err != nil {
				return err
			}
			// stop process
			return nil
		} else {
			log.Print("hash was not equal deleting file from target folder")
			// remove destination file and continue
			if err := removeFile(targetPath); err != nil {
				return err
			}
		}
	}

	tempFileName := fmt.Sprintf("%s.part", targetPath)
	err = copyFile(sourcePath, tempFileName)
	if err != nil {
		return err
	}

	log.Printf("renaming file temp file to '%s'", targetPath)
	// rename file to actual file name
	err = os.Rename(tempFileName, targetPath)
	if err != nil {
		log.Printf("could not rename file %+v", err)
		return err
	}

	log.Printf("file '%s' moved successfully", fileName)
	// The copy was successful, so now delete the original file
	log.Printf("deleting file '%s' from source folder", fileName)
	if err := removeFile(sourcePath); err != nil {
		log.Printf("could not remove source file: %v", err)
		return err
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	return destFile.Sync()
}

func removeFile(path string) error {
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		err := os.RemoveAll(path)
		if err == nil {
			return nil
		}
		if !os.IsPermission(err) {
			return err
		}
		if i == maxRetries-1 {
			return err
		}
		if runtime.GOOS == "windows" {
			time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
		}
	}
	return fmt.Errorf("failed to remove file after %d attempts", maxRetries)
}
