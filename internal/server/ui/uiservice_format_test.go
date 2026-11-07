package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatDuration_WithHours(t *testing.T) {
	// 1h 2m 3s = 3723s = 3723000ms
	got := formatDuration(3723000)
	assert.Equal(t, "1:02:03", got)
}

func TestFormatDuration_WithoutHours(t *testing.T) {
	// 5m 7s = 307s = 307000ms
	got := formatDuration(307000)
	assert.Equal(t, "5:07", got)
}

func TestFormatDuration_Zero(t *testing.T) {
	got := formatDuration(0)
	assert.Equal(t, "0:00", got)
}

func TestFormatDuration_ExactHour(t *testing.T) {
	// 1h = 3600s = 3600000ms
	got := formatDuration(3600000)
	assert.Equal(t, "1:00:00", got)
}

func TestGetFeedTitleFromPath(t *testing.T) {
	got := getFeedTitleFromPath("/media/channel/episode.mp3")
	assert.Equal(t, "channel", got)
}
