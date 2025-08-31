package database

import (
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
)

type PodcastItem struct {
	ID                     string    `json:"id"` // Unique identifier for the video item
	Title                  string    `json:"title"`
	Description            string    `json:"description"`
	Author                 string    `json:"author"`
	Thumbnail              string    `json:"thumbnail"`
	DurationInMilliseconds int64     `json:"duration_in_milliseconds"` // Duration in seconds
	VideoURL               string    `json:"video_url"`
	AudioFilePath          string    `json:"audio_file_path"` // Path to the downloaded audio file
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

func NewPodcastItem(audioFilePath string) (podcastItem *PodcastItem, err error) {
	audioMetadata, err := mp3joiner.GetFFmpegMetadataTag(audioFilePath)
	if err != nil {
		return nil, err
	}

	lengthInSeconds, err := mp3joiner.GetLengthInSeconds(audioFilePath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(audioFilePath)
	if err != nil {
		return nil, err
	}
	fileNameWithoutExtension := strings.TrimSuffix(fileInfo.Name(), filepath.Ext(fileInfo.Name()))

	title := common.ValueOrDefault(audioMetadata[downloader.Title], fileNameWithoutExtension)
	description := common.ValueOrDefault(audioMetadata[downloader.PodcastDescriptionTag], "")

	uploadTime, err := time.Parse("20060102", audioMetadata[downloader.DateTag])
	if err != nil {
		log.Printf("could not parse date tag, reverting to default. error: %v", err)
		uploadTime = fileInfo.ModTime()
	}

	podcastItem = &PodcastItem{
		ID:                     hashVideoUrl(audioFilePath),
		Title:                  title,
		Description:            description,
		Author:                 audioMetadata[downloader.Artist],
		Thumbnail:              audioMetadata[downloader.ThumbnailUrlTag],
		DurationInMilliseconds: int64(lengthInSeconds * 1000), // Convert seconds to milliseconds
		VideoURL:               audioMetadata[downloader.VideoDownloadLink],
		AudioFilePath:          audioFilePath,
		CreatedAt:              uploadTime,
		UpdatedAt:              time.Now(),
	}

	return podcastItem, err
}

func hashVideoUrl(filename string) string {
	// take an audio file path and hash it to a UUIDv4
	data := []byte(filename)
	hash := md5.Sum(data)
	hashBytes := hash[:]

	uuid := make([]byte, 16)
	copy(uuid, hashBytes)

	// Set version (4) and variant bits according to RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16],
	)
}
