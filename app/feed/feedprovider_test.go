package feed

import (
	"reflect"
	"testing"

	"github.com/gorilla/feeds"
)

func Test_valueOrDefault(t *testing.T) {
	type args struct {
		value        any
		defaultValue any
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "string",
			args: args{
				value:        "foo",
				defaultValue: "bar",
			},
			want: "foo",
		}, {
			name: "string empty",
			args: args{
				value:        "",
				defaultValue: "bar",
			},
			want: "bar",
		}, {
			name: "image",
			args: args{
				value:        &feeds.Image{},
				defaultValue: nil,
			},
			want: &feeds.Image{},
		},{
			name: "nil",
			args: args{
				value:        nil,
				defaultValue: &feeds.Image{},
			},
			want: &feeds.Image{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valueOrDefault(tt.args.value, tt.args.defaultValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("valueOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
