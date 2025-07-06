package feed

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

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

func (fp *FeedService) GetFeeds(host string) (feedCollector []*feeds.Feed, err error) {
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
			feed = fp.createFeed(host, directoryName)
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

	parentDirectory := getParentDirectory(podcastItem.AudioFilePath, fp.coreservice.GetAudioSourceDirectory(), fileinfo.Name())
	link := fp.getFeedItemUrl(host, parentDirectory, fileinfo.Name())

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

func getParentDirectory(audioFilePath string, rootFilePath string, fileName string) string {
	parentDirectory := strings.ReplaceAll(audioFilePath, rootFilePath, "")
	parentDirectory = strings.ReplaceAll(parentDirectory, fileName, "")
	parentDirectory = strings.TrimPrefix(parentDirectory, string(os.PathSeparator))
	return parentDirectory
}

func (fp *FeedService) createFeed(host, author string) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       author,
		Link:        &feeds.Link{Href: fp.getFeedUrl(host, author)},
		Description: fmt.Sprintf("%s %s", defaultDescription, author),
		Author:      &feeds.Author{Name: author},
	}

	return feed
}

func (fp *FeedService) getFeedUrl(host, author string) string {
	urlEncodedTitle := url.PathEscape(author)

	return fmt.Sprintf("%s/%s/%s/%s", host, fp.feedItemPath, urlEncodedTitle, defaultURLSuffix)
}

func (fp *FeedService) getFeedItemUrl(host string, parent_folder string, itemName string) string {
	urlEncodedItemName := url.PathEscape(itemName)
	// remove the suffix from the url
	urlPath := strings.TrimSuffix(fp.getFeedUrl(host, parent_folder), defaultURLSuffix)

	return fmt.Sprintf("%s%s", urlPath, urlEncodedItemName)
}
