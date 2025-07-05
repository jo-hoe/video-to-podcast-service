package feed

import (
	"fmt"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/gorilla/feeds"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"
)

func TestCreateFeed(t *testing.T) {
	defaultAuthor := "John Doe"
	type fields struct {
		feedBasePort string
		feedItemPath string
		feedAuthor   string
		coreService  *core.CoreService
	}
	tests := []struct {
		name   string
		fields fields
		want   *feeds.Feed
	}{
		{
			name: "create feed test",
			fields: fields{
				feedBasePort: "443",
				feedItemPath: "v1/feeds",
				feedAuthor:   defaultAuthor,
				coreService:  core.NewCoreService(&database.MockDatabase{}, "testDir"),
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
				coreservice:  tt.fields.coreService,
				feedBasePort: tt.fields.feedBasePort,
				feedItemPath: tt.fields.feedItemPath,
			}
			if got := fp.createFeed(tt.fields.feedAuthor); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createFeed() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewFeedService(t *testing.T) {
	type args struct {
		coreService  *core.CoreService
		feedBasePort string
		feedItemPath string
	}
	tests := []struct {
		name string
		args args
		want *FeedService
	}{
		{
			name: "init test",
			args: args{
				coreService:  core.NewCoreService(&database.MockDatabase{}, "testDir"),
				feedBasePort: "8080",
				feedItemPath: "v1/path",
			},
			want: &FeedService{
				coreservice:  core.NewCoreService(&database.MockDatabase{}, "testDir"),
				feedBasePort: "8080",
				feedItemPath: "v1/path",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewFeedService(tt.args.coreService, tt.args.feedBasePort, tt.args.feedItemPath)
			if got.feedBasePort != tt.want.feedBasePort || got.feedItemPath != tt.want.feedItemPath {
				t.Errorf("NewFeedService() = %v, want %v", got, tt.want)
			}
			// Compare coreService by pointer address (since DeepEqual will fail on different instances)
			if fmt.Sprintf("%p", got.coreservice) != fmt.Sprintf("%p", tt.want.coreservice) {
				t.Errorf("NewFeedService() coreService pointer mismatch")
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
	testHost := "localhost:8080"
	testFilePath, err := filepath.Abs(filepath.Join("..", "..", "..", "test_assets"))
	if err != nil {
		t.Errorf("Error: %v", err)
	}
	mockDB := &database.MockDatabase{}
	coreService := core.NewCoreService(mockDB, testFilePath)
	fp := &FeedService{
		coreservice:  coreService,
		feedBasePort: "8080",
		feedItemPath: "v1/path",
	}
	t.Run("positive test", func(t *testing.T) {
		got, err := fp.GetFeeds(testHost)
		if err != nil {
			t.Errorf("FeedService.GetFeeds() error = %v, wantErr false", err)
			return
		}
		if got == nil {
			t.Error("FeedService.GetFeeds() got = nil")
		}
	})

	fpInvalid := &FeedService{
		coreservice:  core.NewCoreService(mockDB, "non_existing_dir"),
		feedBasePort: "8080",
		feedItemPath: "v1/path",
	}
	t.Run("non existing directory", func(t *testing.T) {
		_, err := fpInvalid.GetFeeds(testHost)
		if err == nil {
			t.Error("FeedService.GetFeeds() expected error for non existing directory, got nil")
		}
	})
}
