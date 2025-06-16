package common

import (
	"reflect"
	"testing"
)

type Test struct {
	Name string
}

func Test_ValueOrDefault(t *testing.T) {
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
				value:        &Test{},
				defaultValue: nil,
			},
			want: &Test{},
		}, {
			name: "nil",
			args: args{
				value:        nil,
				defaultValue: &Test{},
			},
			want: &Test{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValueOrDefault(tt.args.value, tt.args.defaultValue); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("common.ValueOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}
