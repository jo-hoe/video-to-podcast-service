package filemanagement

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var supportedAudioFileExtensions = map[string]bool{
	"mp3":  true,
	"wav":  true,
	"flac": true,
	"ogg":  true,
	"mpeg": true,
}

func GetAudioFiles(directoryPath string) (result []string, err error) {
	// take an input directory and return all audio files in it
	result = make([]string, 0)
	err = filepath.Walk(directoryPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && isSupportedAudioFile(path) {
			result = append(result, path)
		}
		return nil
	})

	return result, err
}

func isSupportedAudioFile(filePath string) bool {
	// plain file extension check
	fileExtension := strings.ToLower(filepath.Ext(filePath))
	fileExtensionParts := strings.Split(fileExtension, ".")
	if len(fileExtensionParts) != 2 {
		return false
	}
	if !supportedAudioFileExtensions[fileExtensionParts[1]] {
		return false
	}

	// more in depth check of file content
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return false
	}
	// check content type which looks like "audio/mp3"
	contentType := http.DetectContentType(buffer)
	contentTypeParts := strings.Split(contentType, "/")
	if len(contentTypeParts) != 2 {
		return false
	}
	if contentTypeParts[0] != "audio" {
		return false
	}
	if !supportedAudioFileExtensions[contentTypeParts[1]] {
		return false
	}

	return true
}
