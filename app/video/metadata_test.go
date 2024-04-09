package video

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetMetadata(t *testing.T) {
	rootDirectory, err := os.MkdirTemp(os.TempDir(), "testDir")
	defer os.RemoveAll(rootDirectory)
	if err != nil {
		t.Error("could not create folder")
	}

	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "assets", "test", "video.mp4"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	type args struct {
		inputFilePath string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "get metadata",
			args: args{
				inputFilePath: testFilePath,
			},
			want: "---",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetMetadata(tt.args.inputFilePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}
