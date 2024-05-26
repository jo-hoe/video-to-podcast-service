package download

import (
	"reflect"
	"testing"
)

func Test_sanitizeFilename(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nothing to change",
			args: args{
				filename: "test",
			},
			want: "test",
		},
		{
			name: "remove invalid characters",
			args: args{
				filename: "test?",
			},
			want: "test_",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilename(tt.args.filename); got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetVideoDownloader(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name           string
		args           args
		wantDownloader AudioDownloader
		wantErr        bool
	}{
		{
			name: "youtube",
			args: args{
				url: "https://www.youtube.com/watch?v=dQw4w9WgXcQ",
			},
			wantDownloader: &YoutubeAudioDownloader{},
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
