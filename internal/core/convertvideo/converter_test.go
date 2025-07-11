package convertvideo

import (
	"os"
	"path/filepath"
	"testing"
)

func Test_convertVideoToAudio(t *testing.T) {
	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	defer func() {
		if err := os.RemoveAll(rootDirectory); err != nil {
			t.Errorf("Error removing temp directory: %v", err)
		}
	}()
	if err != nil {
		t.Error("could not create folder")
	}

	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "..", "test_assets", "video.mp4"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	type args struct {
		inputFilePath  string
		outputFilePath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Convert a video to audio file successfully",
			args: args{
				inputFilePath:  testFilePath,
				outputFilePath: filepath.Join(rootDirectory, "output.mp3"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ConvertVideoToAudio(tt.args.inputFilePath, tt.args.outputFilePath); (err != nil) != tt.wantErr {
				t.Errorf("ConvertVideoToAudio() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
