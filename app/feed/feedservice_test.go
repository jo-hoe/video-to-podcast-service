package feed

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gorilla/feeds"
)

func TestCreateFeed(t *testing.T) {
	defaultAuthor := "John Doe"
	type fields struct {
		feedBaseUrl          string
		feedBasePort         string
		feedItemPath         string
		feedAuthor           string
		audioSourceDirectory string
	}
	tests := []struct {
		name   string
		fields fields
		want   *feeds.Feed
	}{
		{
			name: "create feed test",
			fields: fields{
				feedBaseUrl:          "https://example.com",
				feedBasePort:         "443",
				feedItemPath:         "v1/feeds",
				feedAuthor:           defaultAuthor,
				audioSourceDirectory: "",
			},
			want: &feeds.Feed{
				Title:       defaultAuthor,
				Link:        &feeds.Link{Href: "https://example.com:443/v1/feeds/John%20Doe/rss.xml"},
				Description: fmt.Sprintf("%s %s", defaultDescription, defaultAuthor),
				Author:      &feeds.Author{Name: defaultAuthor},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &FeedService{
				feedBaseUrl:          tt.fields.feedBaseUrl,
				feedBasePort:         tt.fields.feedBasePort,
				audioSourceDirectory: tt.fields.audioSourceDirectory,
				feedItemPath:         tt.fields.feedItemPath,
			}
			if got := fp.createFeed(tt.fields.feedAuthor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFeedService(t *testing.T) {
	type args struct {
		audioSourceDirectory string
		feedBaseUrl          string
		feedBasePort         string
		feedItemPath         string
	}
	tests := []struct {
		name string
		args args
		want *FeedService
	}{
		{
			name: "init test",
			args: args{
				audioSourceDirectory: "testDir",
				feedBaseUrl:          "testUrl",
				feedBasePort:         "8080",
				feedItemPath:         "v1/path",
			},
			want: &FeedService{
				audioSourceDirectory: "testDir",
				feedBaseUrl:          "testUrl",
				feedBasePort:         "8080",
				feedItemPath:         "v1/path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeedService(tt.args.audioSourceDirectory, tt.args.feedBaseUrl, tt.args.feedBasePort, tt.args.feedItemPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFeedService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeedService_GetFeeds(t *testing.T) {
	defaultURL := "127.0.0.1"
	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "assets", "test"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	tests := []struct {
		name    string
		fp      *FeedService
		wantErr bool
	}{
		{
			name: "positive test",
			fp: &FeedService{
				feedBaseUrl:          defaultURL,
				audioSourceDirectory: testFilePath,
			},
			wantErr: false,
		},
		{
			name: "non existing directory",
			fp: &FeedService{
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
				t.Errorf("FeedService.GetFeeds() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil && !tt.wantErr {
				t.Error("FeedService.GetFeeds() got = nil")
			}
			if len(got) != 3 && !tt.wantErr {
				t.Errorf("expected 3 feeds, got %d", len(got))
			}
		})
	}
}
