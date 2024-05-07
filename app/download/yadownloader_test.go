package download

import (
	"testing"
)

func Test_getYoutubeVideoId(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "Positive test case",
			args: args{
				url: "https://www.youtube.com/watch?v=BaW_jenozKc",
			},
			want:    "BaW_jenozKc",
			wantErr: false,
		},
		{
			name: "Negative test case",
			args: args{
				url: "garbage",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "Short link",
			args: args{
				url: "https://youtu.be/DucriSA8ukw?feature=shared",
			},
			want:    "DucriSA8ukw",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getVideoId(tt.args.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVideoId() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getVideoId() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYoutubeAudioDownloader_IsSupported(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name string
		y    *YoutubeAudioDownloader
		args args
		want bool
	}{
		{
			name: "test video link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://www.youtube.com/watch?v=jNQXAC9IVRw&pp=ygUQb25lIHNlY29uZCB2aWRlbw%3D%3D",
			},
			want: true,
		},
		{
			name: "test playlist link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://www.youtube.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-",
			},
			want: true,
		},
		{
			name: "test short video link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://youtu.be/DucriSA8ukw?feature=shared",
			},
			want: true,
		},
		{
			name: "test not existing link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://not-existing.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-",
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.y.IsSupported(tt.args.url); got != tt.want {
				t.Errorf("YoutubeAudioDownloader.IsSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}
