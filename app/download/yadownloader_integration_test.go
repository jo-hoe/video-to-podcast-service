package download

import (
	"os"
	"path/filepath"
	"testing"
)

func TestYoutubeAudioDownloader_Download(t *testing.T) {
	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	defer os.RemoveAll(rootDirectory)
	if err != nil {
		t.Error("could not create folder")
	}

	type args struct {
		urlString string
		path      string
	}
	tests := []struct {
		name    string
		y       *YoutubeAudioDownloader
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "Video Download Test",
			y:    &YoutubeAudioDownloader{},
			args: args{
				urlString: "https://www.youtube.com/watch?v=jNQXAC9IVRw&pp=ygUQb25lIHNlY29uZCB2aWRlbw%3D%3D",
				path:      rootDirectory,
			},
			want:    []string{filepath.Join(rootDirectory, "Me at the zoo.mp4")},
			wantErr: false,
		},
		{
			name: "Playlist Download Test",
			y:    &YoutubeAudioDownloader{},
			args: args{
				urlString: "https://www.youtube.com/playlist?list=PLXqZLJI1Rpy_x_piwxi9T-UlToz3UGdM-",
				path:      rootDirectory,
			},
			want:    make([]string, 10),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.y.Download(tt.args.urlString, tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("YoutubeAudioDownloader.Download() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("YoutubeAudioDownloader.Download() = %v, want %v", got, tt.want)
			}
		})
	}
}
