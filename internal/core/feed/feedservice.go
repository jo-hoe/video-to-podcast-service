package feed

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"

	"github.com/jo-hoe/gofeedx"
)

const (
	defaultURLSuffix   = "rss.xml"
	defaultDescription = "Podcast Feed of"
	mp3KeyAttribute    = "artist"
)

type FeedService struct {
	coreservice  *core.CoreService
	feedBasePort string
	feedItemPath string
}

func NewFeedService(
	coreService *core.CoreService,
	feedBasePort string,
	feedItemPath string) *FeedService {
	return &FeedService{
		coreservice:  coreService,
		feedBasePort: feedBasePort,
		feedItemPath: feedItemPath,
	}
}

func (fp *FeedService) GetFeeds(host string) (feedCollector []*gofeedx.Feed, err error) {
	feedCollector = make([]*gofeedx.Feed, 0)

	podcastItems, err := fp.coreservice.GetDatabaseService().GetAllPodcastItems()
	if err != nil {
		return nil, fmt.Errorf("could not get podcast items: %w", err)
	}

	for _, podcastItem := range podcastItems {
		// create feed item from podcast item
		item, err := fp.createFeedItem(host, podcastItem)
		if err != nil {
			return nil, fmt.Errorf("could not create feed item: %w", err)
		}

		directoryName := filepath.Base(filepath.Dir(podcastItem.AudioFilePath))
		feed := fp.getFeedWithAuthor(directoryName, feedCollector)
		if feed == nil {
			feed = fp.createFeed(host, directoryName, podcastItem.AudioFilePath)
			feedCollector = append(feedCollector, feed)
		}
		feed.Items = append(feed.Items, item)

		// inject image if not already set, using the thumbnail of the podcast item
		if feed.Image == nil {
			feed.Image = &gofeedx.Image{
				Url:   podcastItem.Thumbnail,
				Link:  podcastItem.Thumbnail,
				Title: directoryName,
			}
		}
	}

	return feedCollector, nil
}

func (fp *FeedService) getFeedWithAuthor(author string, feeds []*gofeedx.Feed) *gofeedx.Feed {
	for _, feed := range feeds {
		if feed.Author != nil && feed.Author.Name == author {
			return feed
		}
	}
	return nil
}

func (fp *FeedService) createFeedItem(host string, podcastItem *database.PodcastItem) (*gofeedx.Item, error) {
	fileinfo, err := os.Stat(podcastItem.AudioFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not get file info for %s: %w", podcastItem.AudioFilePath, err)
	}
	if fileinfo.IsDir() {
		return nil, fmt.Errorf("expected file but got directory: %s", podcastItem.AudioFilePath)
	}

	link := fp.coreservice.GetLinkToAudioFile(host, fp.feedItemPath, podcastItem.AudioFilePath)

	itemBuilder := gofeedx.NewItem(escapeXML(podcastItem.Title)).
		WithGUID(podcastItem.ID, "false").
		WithLink(link).
		WithDescription(escapeXML(podcastItem.Description)).
		WithAuthor(common.ValueOrDefault(podcastItem.Author, ""), "").
		WithCreated(podcastItem.CreatedAt).
		WithUpdated(podcastItem.UpdatedAt).
		WithEnclosure(link, fileinfo.Size(), "audio/mpeg").
		WithDurationSeconds(int(podcastItem.DurationInMilliseconds / 1000)).
		WithPSPImageHref(podcastItem.Thumbnail)

	return itemBuilder.Build()
}

func escapeXML(s string) string {
	// gofeedx should handle escaping, but we can add additional escaping if needed
	return fmt.Sprintf("<![CDATA[%s]]>", s)
}

func (fp *FeedService) createFeed(host string, author string, filepath string) *gofeedx.Feed {
	selfURL := fp.coreservice.GetLinkToFeed(host, fp.feedItemPath, filepath)

	feedBuilder := gofeedx.NewFeed(author).
		WithLink(selfURL).
		WithDescription(fmt.Sprintf("%s %s", defaultDescription, author)).
		WithAuthor(author, "").
		WithFeedURL(selfURL)

	built, err := feedBuilder.Build()
	if err != nil {
		slog.Error("failed to build feed", "error", err)
		return nil
	}
	return built
}
