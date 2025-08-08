package feed

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jo-hoe/video-to-podcast-service/internal/config"
	"github.com/jo-hoe/video-to-podcast-service/internal/core"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/common"
	"github.com/jo-hoe/video-to-podcast-service/internal/core/database"

	"github.com/gorilla/feeds"
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
	feedConfig   *config.FeedConfig
}

func NewFeedService(
	coreService *core.CoreService,
	feedBasePort string,
	feedItemPath string,
	feedConfig *config.FeedConfig) *FeedService {
	return &FeedService{
		coreservice:  coreService,
		feedBasePort: feedBasePort,
		feedItemPath: feedItemPath,
		feedConfig:   feedConfig,
	}
}

func (fp *FeedService) GetFeeds(host string) (feedCollector []*feeds.Feed, err error) {
	if fp.feedConfig != nil && fp.feedConfig.Mode == "unified" {
		feed, err := fp.getUnifiedFeed(host)
		if err != nil {
			return nil, err
		}
		return []*feeds.Feed{feed}, nil
	}

	// Default to per-directory mode
	return fp.getPerDirectoryFeeds(host)
}

func (fp *FeedService) getPerDirectoryFeeds(host string) (feedCollector []*feeds.Feed, err error) {
	feedCollector = make([]*feeds.Feed, 0)

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

		if feed.Image == nil {
			feed.Image = &feeds.Image{
				Url:  podcastItem.Thumbnail,
				Link: podcastItem.Thumbnail,
			}
		}
	}

	return feedCollector, nil
}

func (fp *FeedService) getUnifiedFeed(host string) (*feeds.Feed, error) {
	podcastItems, err := fp.coreservice.GetDatabaseService().GetAllPodcastItems()
	if err != nil {
		return nil, fmt.Errorf("could not get podcast items: %w", err)
	}

	// Create unified feed
	feed := &feeds.Feed{
		Title:       "All Podcast Items",
		Link:        &feeds.Link{Href: fmt.Sprintf("http://%s%s/v1/feeds/all/rss.xml", host, fp.feedBasePort)},
		Description: "Unified podcast feed containing all items",
		Author:      &feeds.Author{Name: "Video to Podcast Service"},
	}

	// Add all items to the unified feed
	for _, podcastItem := range podcastItems {
		item, err := fp.createFeedItem(host, podcastItem)
		if err != nil {
			return nil, fmt.Errorf("could not create feed item: %w", err)
		}
		feed.Items = append(feed.Items, item)

		// Set feed image from first item with thumbnail
		if feed.Image == nil && podcastItem.Thumbnail != "" {
			feed.Image = &feeds.Image{
				Url:  podcastItem.Thumbnail,
				Link: podcastItem.Thumbnail,
			}
		}
	}

	// Sort items by creation date for consistent ordering (newest first)
	for i := 0; i < len(feed.Items)-1; i++ {
		for j := i + 1; j < len(feed.Items); j++ {
			if feed.Items[i].Created.Before(feed.Items[j].Created) {
				feed.Items[i], feed.Items[j] = feed.Items[j], feed.Items[i]
			}
		}
	}

	return feed, nil
}

func (fp *FeedService) getFeedWithAuthor(author string, feeds []*feeds.Feed) *feeds.Feed {
	for _, feed := range feeds {
		if feed.Author.Name == author {
			return feed
		}
	}
	return nil
}

func (fp *FeedService) createFeedItem(host string, podcastItem *database.PodcastItem) (*feeds.Item, error) {
	fileinfo, err := os.Stat(podcastItem.AudioFilePath)
	if err != nil {
		return nil, fmt.Errorf("could not get file info for %s: %w", podcastItem.AudioFilePath, err)
	}
	if fileinfo.IsDir() {
		return nil, fmt.Errorf("expected file but got directory: %s", podcastItem.AudioFilePath)
	}

	link := fp.coreservice.GetLinkToAudioFile(host, fp.feedItemPath, podcastItem.AudioFilePath)

	return &feeds.Item{
		Id:          podcastItem.ID,
		Title:       podcastItem.Title,
		Link:        &feeds.Link{Href: link},
		Description: podcastItem.Description,
		Author:      &feeds.Author{Name: common.ValueOrDefault(podcastItem.Author, "")},
		Created:     podcastItem.CreatedAt,
		IsPermaLink: "false",
		Enclosure: &feeds.Enclosure{
			Url:    link,
			Type:   "audio/mpeg",
			Length: fmt.Sprintf("%d", podcastItem.DurationInMilliseconds),
		},
	}, nil
}

func (fp *FeedService) createFeed(host string, author string, filepath string) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       author,
		Link:        &feeds.Link{Href: fp.coreservice.GetLinkToFeed(host, fp.feedItemPath, filepath)},
		Description: fmt.Sprintf("%s %s", defaultDescription, author),
		Author:      &feeds.Author{Name: author},
	}

	return feed
}
