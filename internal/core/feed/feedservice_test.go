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
		feedBasePort      string
		feedItemPath      string
		feedAudioFilePath string
		feedAuthor        string
		feedHost          string
		coreService       *core.CoreService
	}
	tests := []struct {
		name   string
		fields fields
		want   *feeds.Feed
	}{
		{
			name: "create feed test",
			fields: fields{
				feedBasePort:      "443",
				feedItemPath:      "v1/feeds",
				feedAuthor:        defaultAuthor,
				feedHost:          "localhost",
				feedAudioFilePath: filepath.Join("c", "testDir", "audio.mp3"),
				coreService:       core.NewCoreService(&database.MockDatabase{}, filepath.Join("c"), nil),
			},
			want: &feeds.Feed{
				Title:       defaultAuthor,
				Link:        &feeds.Link{Href: "http://localhost/v1/feeds/testDir/rss.xml"},
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
			if got := fp.createFeed(tt.fields.feedHost, tt.fields.feedAuthor, tt.fields.feedAudioFilePath); !reflect.DeepEqual(got, tt.want) {
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
	sharedCore := core.NewCoreService(&database.MockDatabase{}, "testDir", nil)
	tests := []struct {
		name string
		args args
		want *FeedService
	}{
		{
			name: "init test",
			args: args{
				coreService:  sharedCore,
				feedBasePort: "8080",
				feedItemPath: "v1/path",
			},
			want: &FeedService{
				coreservice:  sharedCore,
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
			if got.coreservice != tt.want.coreservice {
				t.Errorf("NewFeedService() coreService pointer mismatch")
			}
		})
	}
}
