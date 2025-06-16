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
				Link:        &feeds.Link{Href: "/v1/feeds/John%20Doe/rss.xml"},
				Description: fmt.Sprintf("%s %s", defaultDescription, defaultAuthor),
				Author:      &feeds.Author{Name: defaultAuthor},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fp := &FeedService{
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
				feedBasePort:         "8080",
				feedItemPath:         "v1/path",
			},
			want: &FeedService{
				audioSourceDirectory: "testDir",
				feedBasePort:         "8080",
				feedItemPath:         "v1/path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewFeedService(tt.args.audioSourceDirectory, tt.args.feedBasePort, tt.args.feedItemPath); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewFeedService() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_hashFileNameToUUIDv4(t *testing.T) {
	result := hashFileNameToUUIDv4("my_demo_file.mp3")
	result2 := hashFileNameToUUIDv4("my_demo_file'.mp3")

	//the that result is a valid UUIDv4
	if result == "" {
		t.Errorf("hashFileNameToUUIDv4() returned empty string")
	}

	if len(result) != 36 {
		t.Errorf("hashFileNameToUUIDv4() returned string of length %d, want 36", len(result))
	}

	// Basic UUIDv4 format check: xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx
	if result[14] != '4' {
		t.Errorf("hashFileNameToUUIDv4() returned string with wrong version: %s", result)
	}
	if result[19] != '8' && result[19] != '9' && result[19] != 'a' && result[19] != 'b' {
		t.Errorf("hashFileNameToUUIDv4() returned string with wrong variant: %s", result)
	}

	if result2 == result {
		t.Errorf("hashFileNameToUUIDv4() returned same UUID for different input: %s", result)
	}
}

func TestFeedService_GetFeeds(t *testing.T) {
	testFilePath, err := filepath.Abs(filepath.Join("..", "..","..", "test_assets"))
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
				audioSourceDirectory: testFilePath,
			},
			wantErr: false,
		},
		{
			name: "non existing directory",
			fp: &FeedService{
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
			if len(got) != 2 && !tt.wantErr {
				t.Errorf("expected 2 feeds, got %d", len(got))
			}
		})
	}
}
