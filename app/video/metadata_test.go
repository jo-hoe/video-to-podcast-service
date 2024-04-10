package video

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetThumbnailPicture(t *testing.T) {
	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	defer os.RemoveAll(rootDirectory)
	if err != nil {
		t.Error("could not create folder")
	}

	videoFilePath, err := filepath.Abs(filepath.Join("..", "..", "assets", "test", "video.mp4"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	imageFilePath, err := filepath.Abs(filepath.Join("..", "..", "assets", "test", "image.jpg"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	type args struct {
		mediaFilePath string
		imageFilePath string
		outputFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "set thumbnail picture to video file",
			args: args{
				mediaFilePath: videoFilePath,
				imageFilePath: imageFilePath,
				outputFilePath: filepath.Join(rootDirectory, "output.mp4"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := SetThumbnailPicture(tt.args.mediaFilePath, tt.args.imageFilePath, tt.args.outputFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetThumbnailPicture() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
