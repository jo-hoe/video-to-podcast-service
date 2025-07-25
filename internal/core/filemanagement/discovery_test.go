package filemanagement

import (
	"path/filepath"
	"reflect"
	"testing"
)

func Test_GetAudioFiles(t *testing.T) {
	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "..", "test", "assets"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	type args struct {
		directoryPath string
	}
	tests := []struct {
		name       string
		args       args
		wantResult []string
		wantErr    bool
	}{
		{
			name: "Find MP3 files",
			args: args{
				directoryPath: testFilePath,
			},
			wantResult: []string{
				filepath.Join(testFilePath, "audio11.mp3"),
				filepath.Join(testFilePath, "audio12.mp3"),
				filepath.Join(testFilePath, "audio21.mp3"),
			},
			wantErr: false,
		},
		{
			name: "Non Existing Directory",
			args: args{
				directoryPath: "test/non-existing-directory",
			},
			wantResult: make([]string, 0),
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := GetAudioFiles(tt.args.directoryPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetAudioFiles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("GetAudioFiles() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
