package discovery

import (
	"path/filepath"
	"reflect"
	"testing"
)


func Test_GetAudioFiles(t *testing.T) {
	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "assets", "test"))
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
			wantResult: []string{filepath.Join(testFilePath, "audio.mp3")},
			wantErr: false,
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
