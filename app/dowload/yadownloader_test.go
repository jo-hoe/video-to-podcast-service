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