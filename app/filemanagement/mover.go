package filemanagement

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
)

func calculateFileHash(filePath string) ([]byte, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return make([]byte, 0), err
	}
	defer file.Close()

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
			log.Print("hash was not equal deleting file from target folder")
			// remove destination file and continue
			err = os.Remove(targetPath)
			if err != nil {
				return err
			}
		}
	}

	outputFile, err := os.Create(targetPath)
	if err != nil {
		inputFile.Close()
		return err
	}
	defer func() {
		log.Print("securely cleaning cache and closing file")
		fileClosingError := outputFile.Close()
		if fileClosingError != nil {
			log.Print("could not close file")
			return
		}

		// check if copying was successful
		if err != nil {
			return
		}

		// currently double-check of file hash
		// target/source file has been omitted

		log.Printf("file '%s' moved successfully", fileName)
		// The copy was successful, so now delete the original file
		log.Printf("deleting file '%s' from source folder", fileName)
		err = os.Remove(sourcePath)
		if err != nil {
			log.Printf("could not close file %+v", err)
		}

	}()

	// actual file copy
	_, err = io.Copy(outputFile, inputFile)
	inputFile.Close()
	if err != nil {
		return err
	}

	return nil
}
