package feed

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/gorilla/feeds"
)

func TestCreateFeed(t *testing.T) {
	defaultAuthor := "John Doe"
	type fields struct {
		feedBaseUrl          string
		feedBasePort        string
		feedAuthor           string
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
				feedBaseUrl:          "",
				feedBasePort:         "",
				feedAuthor:           defaultAuthor,
				audioSourceDirectory: "",
			},
			want: &feeds.Feed{
				Title:       fmt.Sprintf("%s %s", defaultTitlePrefix, defaultAuthor),
				Link:        &feeds.Link{Href: "127.0.0.1:8080/feeds/John%20Doe/rss.xml"},
				Description: fmt.Sprintf("%s %s", defaultDescription, defaultAuthor),
				Author:      &feeds.Author{Name: defaultAuthor},
				Updated:     time.Time{},
			},
		},
		{
			name: "custom values",
			fields: fields{
				feedBaseUrl:          "https://example.com",
				feedBasePort:         "443",
				feedAuthor:           defaultAuthor,
				audioSourceDirectory: "",
			},
			want: &feeds.Feed{
				Title:       fmt.Sprintf("%s %s", defaultTitlePrefix, defaultAuthor),
				Link:        &feeds.Link{Href: "https://example.com:443/feeds/John%20Doe/rss.xml"},
				Description: fmt.Sprintf("%s %s", defaultDescription, defaultAuthor),
				Author:      &feeds.Author{Name: defaultAuthor},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &FeedProvider{
				feedBaseUrl:          tt.fields.feedBaseUrl,
				feedBasePort:         tt.fields.feedBasePort,
				audioSourceDirectory: tt.fields.audioSourceDirectory,
			}
			if got := fp.createFeed(tt.fields.feedAuthor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFeedProvider(t *testing.T) {
	type args struct {
		audioSourceDirectory string
		feedBaseUrl          string
		feedBasePort         string
	}
	tests := []struct {
		name string
		args args
		want *FeedProvider
	}{
		{
			name: "init test",
			args: args{
				audioSourceDirectory: "testDir",
				feedBaseUrl:          "testUrl",
				feedBasePort:         "8080",
			},
			want: &FeedProvider{
				audioSourceDirectory: "testDir",
				feedBaseUrl:          "testUrl",
				feedBasePort:         "8080",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeedProvider(tt.args.audioSourceDirectory, tt.args.feedBaseUrl, tt.args.feedBasePort); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFeedProvider() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeedProvider_GetFeeds(t *testing.T) {
	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "assets", "test"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	tests := []struct {
		name    string
		fp      *FeedProvider
		wantErr bool
	}{
		{
			name: "positive test",
			fp: &FeedProvider{
				feedBaseUrl:          defaultURL,
				audioSourceDirectory: testFilePath,
			},
			wantErr: false,
		},
		{
			name: "non existing directory",
			fp: &FeedProvider{
				feedBaseUrl:          defaultURL,
				audioSourceDirectory: "testDir",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fp.GetFeeds()
			if (err != nil) != tt.wantErr {
				t.Errorf("FeedProvider.GetFeeds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil && !tt.wantErr {
				t.Error("FeedProvider.GetFeeds() got = nil")
			}
			if len(got) != 2 && !tt.wantErr {
				t.Errorf("expected 2 feeds, got %d", len(got))
			}
		})
	}
}
