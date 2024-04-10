package video

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGetTagMetadata(t *testing.T) {
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
		path string
	}
	tests := []struct {
		name       string
		args       args
		wantResult map[string]string
		wantErr    bool
	}{
		{
			name: "positive test",
			args: args{
				path: testFilePath,
			},
			wantResult: map[string]string{
				"comment":           "https://www.youtube.com/watch?v=thisisatest",
				"major_brand":       "isom",
				"minor_version":     "512",
				"compatible_brands": "isomiso2avc1mp41",
				"encoder":           "Lavf60.3.100",
				"date":              "2024",
				"title":             "Test Title",
				"artist":            "Tester",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := GetTagMetadata(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetTagMetadata() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetTagMetadata() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}

func TestSetTagMetadata(t *testing.T) {
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
		inputPath  string
		tags       map[string]string
		outputPath string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "positive test",
			args: args{
				inputPath:  testFilePath,
				tags:       map[string]string{"title": "testtitle", "artist": "testartist"},
				outputPath: filepath.Join(rootDirectory, "output.mp4"),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetTagMetadata(tt.args.inputPath, tt.args.tags, tt.args.outputPath); (err != nil) != tt.wantErr {
				t.Errorf("SetTagMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
