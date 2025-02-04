package youtube

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"slices"

	mp3joiner "github.com/jo-hoe/mp3-joiner"
	"github.com/jo-hoe/video-to-podcast-service/app/filemanagement"
)

type SuperBlockResponse struct {
	Category      string    `json:"category"`
	ActionType    string    `json:"actionType"`
	Segment       []float64 `json:"segment"`
	UUID          string    `json:"UUID"`
	VideoDuration float64   `json:"videoDuration"`
	Locked        int       `json:"locked"`
	Votes         int       `json:"votes"`
	Description   string    `json:"description"`
}

const (
	baseURL       = "https://sponsor.ajay.app/api/skipSegments"
	queryTemplate = "?videoID=%s&categories=[\"sponsor\",\"selfpromo\",\"interaction\",\"intro\",\"outro\",\"preview\",\"music_offtopic\",\"filler\"]"
	videoIdKey    = ""
)

func RemoveAds(httpClient http.Client, filePath string) error {
	videoId, err := getVideoId(filePath)
	if err != nil {
		return err
	}

	allSegments, err := getSuperBlockResponse(httpClient, videoId)
	if err != nil {
		return err
	}

	// filter out item without segments
	segmentsToRemove := slices.Collect(
		func(yield func(SuperBlockResponse) bool) {
			for _, item := range allSegments {
				if len(item.Segment) > 0 {
					if !yield(item) {
						return
					}
				}
			}
		},
	)

	if len(segmentsToRemove) < 1 {
		log.Printf("no segments to remove found for '%s'", filePath)
		return err
	}

	// Sort segments by start time
	slices.SortFunc(segmentsToRemove, func(a, b SuperBlockResponse) int {
		if a.Segment[0] < b.Segment[0] {
			return -1
		}
		if a.Segment[0] > b.Segment[0] {
			return 1
		}
		return 0
	})

	// get video duration
	videoDuration := segmentsToRemove[0].VideoDuration
	// create segments to keep (invert the ad segments)
	segmentsToAddToAudioFile := make([][]float64, 2)
	currentTime := 0.0

	for _, segment := range segmentsToRemove {
		// add segment from current time to start of ad
		if segment.Segment[0] > currentTime {
			segmentsToAddToAudioFile = append(segmentsToAddToAudioFile, []float64{currentTime, segment.Segment[0]})
		}
		currentTime = segment.Segment[1]
	}

	// add final segment if there's remaining content
	if currentTime < videoDuration {
		segmentsToAddToAudioFile = append(segmentsToAddToAudioFile, []float64{currentTime, videoDuration})
	}

	builder := mp3joiner.NewMP3Builder()
	for _, segment := range segmentsToAddToAudioFile {
		builder.Append(filePath, segment[0], segment[1])
	}
	tempFilename := fmt.Sprintf("%s_temp", filePath)
	err = builder.Build(tempFilename)
	if err != nil {
		return err
	}

	return filemanagement.MoveFile(tempFilename, filePath)
}

func getVideoId(filePath string) (result string, err error) {
	metadata, err := mp3joiner.GetFFmpegMetadataTag(filePath)
	if err != nil {
		return result, err
	}

	return metadata[videoIdKey], err
}

func getSuperBlockResponse(httpClient http.Client, videoId string) (result []SuperBlockResponse, err error) {
	queryString := fmt.Sprintf(queryTemplate, videoId)
	url := fmt.Sprintf("%s%s", baseURL, queryString)
	response, err := httpClient.Get(url)
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, err
}
