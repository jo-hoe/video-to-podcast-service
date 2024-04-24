package download

import "testing"

func Test_sanitizeFilename(t *testing.T) {
	type args struct {
		filename string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "nothing to change",
			args: args{
				filename: "test",
			},
			want: "test",
		},
		{
			name: "remove invalid characters",
			args: args{
				filename: "test?",
			},
			want: "test_",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizeFilename(tt.args.filename); got != tt.want {
				t.Errorf("sanitizeFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}
