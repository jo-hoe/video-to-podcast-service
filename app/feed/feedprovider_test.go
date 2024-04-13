package feed

import (
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/feeds"
)

func TestCreateFeed(t *testing.T) {
	defaultTime := time.Now()
	type fields struct {
		feedTitle            string
		feedBaseUrl          string
		feedDescription      string
		feedAuthor           string
		feedCreated          time.Time
		feedImage            *feeds.Image
		audioSourceDirectory string
	}
	tests := []struct {
		name   string
		fields fields
		want   *feeds.Feed
	}{
		{
			name: "default values",
			fields: fields{
				feedTitle:            "",
				feedBaseUrl:          "",
				feedDescription:      "",
				feedAuthor:           "",
				feedCreated:          defaultTime,
				feedImage:            nil,
				audioSourceDirectory: "",
			},
			want: &feeds.Feed{
				Title:       "Rss Feed",
				Link:        &feeds.Link{Href: "127.0.0.1:8080/rss.xml"},
				Description: "",
				Author:      &feeds.Author{Name: ""},
				Created:     defaultTime,
				Image:       nil,
			},
		},
		{
			name: "custom values",
			fields: fields{
				feedTitle:            "My Feed Title",
				feedBaseUrl:          "https://example.com/feed",
				feedDescription:      "This is my feed description",
				feedAuthor:           "John Doe",
				feedCreated:          defaultTime,
				feedImage:            &feeds.Image{Url: "https://example.com/image.png"}, // TODO: add more image fields
				audioSourceDirectory: "",
			},
			want: &feeds.Feed{
				Title:       "My Feed Title",
				Link:        &feeds.Link{Href: "https://example.com/feed"},
				Description: "This is my feed description",
				Author:      &feeds.Author{Name: "John Doe"},
				Created:     defaultTime,
				Image:       &feeds.Image{Url: "https://example.com/image.png"}, // TODO: add more image fields
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &FeedProvider{
				feedTitle:            tt.fields.feedTitle,
				feedBaseUrl:          tt.fields.feedBaseUrl,
				feedDescription:      tt.fields.feedDescription,
				feedAuthor:           tt.fields.feedAuthor,
				feedCreated:          tt.fields.feedCreated,
				feedImage:            tt.fields.feedImage,
				audioSourceDirectory: tt.fields.audioSourceDirectory,
			}
			if got := fp.createFeed(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

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
		}, {
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
