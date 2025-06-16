package download

import (
	"reflect"
	"testing"

	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/downloader"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/download/youtube"
)

func TestGetVideoDownloader(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name           string
		args           args
		wantDownloader downloader.AudioDownloader
		wantErr        bool
	}{
		{
			name: "youtube",
			args: args{
				url: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			},
			wantDownloader: &youtube.YoutubeAudioDownloader{},
			wantErr:        false,
		},
		{
			name: "unsupport url",
			args: args{
				url: "https://unsupport.com/123456789",
			},
			wantDownloader: nil,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotDownloader, err := GetVideoDownloader(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVideoDownloader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotDownloader, tt.wantDownloader) {
				t.Errorf("GetVideoDownloader() = %v, want %v", gotDownloader, tt.wantDownloader)
			}
		})
	}
}
