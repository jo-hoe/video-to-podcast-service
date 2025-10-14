package youtube

import (
	"testing"
)

func TestYoutubeAudioDownloader_IsVideoSupported(t *testing.T) {
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
			name: "test shorter video link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://youtu.be/DucriSA8ukw?feature=shared",
			},
			want: true,
		},
		{
			name: "test youtube shorts link",
			y:    &YoutubeAudioDownloader{},
			args: args{
				url: "https://www.youtube.com/shorts/Hb3rmh-_FMw",
			},
			want: false,
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
			if got := tt.y.IsVideoSupported(tt.args.url); got != tt.want {
				t.Errorf("YoutubeAudioDownloader.IsVideoSupported() = %v, want %v", got, tt.want)
			}
		})
	}
}
