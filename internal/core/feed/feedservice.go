package feed

import (
	"crypto/md5"
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
			feed = fp.createFeed(directoryName)
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

func hashFileNameToUUIDv4(filename string) string {
	// take an audio file path and hash it to a UUIDv4
	data := []byte(filename)
	hash := md5.Sum(data)
	hashBytes := hash[:]

	uuid := make([]byte, 16)
	copy(uuid, hashBytes)

	// Set version (4) and variant bits according to RFC 4122
	uuid[6] = (uuid[6] & 0x0f) | 0x40 // Version 4
	uuid[8] = (uuid[8] & 0x3f) | 0x80 // Variant is 10

	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		uuid[0:4],
		uuid[4:6],
		uuid[6:8],
		uuid[8:10],
		uuid[10:16],
	)
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

	return &feeds.Item{
		Id:          podcastItem.ID,
		Title:       podcastItem.Title,
		Link:        &feeds.Link{Href: fp.getFeedItemUrl(host, parentDirectory, fileinfo.Name())},
		Description: podcastItem.Description,
		Author:      &feeds.Author{Name: common.ValueOrDefault(podcastItem.Author, "")},
		Created:     podcastItem.CreatedAt,
		IsPermaLink: "false",
	}, nil
}

func getParentDirectory(audioFilePath string, rootFilePath string, fileName string) string {
	parentDirectory := strings.ReplaceAll(audioFilePath, rootFilePath, "")
	parentDirectory = strings.ReplaceAll(parentDirectory, fileName, "")
	parentDirectory = strings.TrimPrefix(parentDirectory, string(os.PathSeparator))
	return parentDirectory
}

func (fp *FeedService) createFeed(author string) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       author,
		Link:        &feeds.Link{Href: fp.getFeedUrl(author)},
		Description: fmt.Sprintf("%s %s", defaultDescription, author),
		Author:      &feeds.Author{Name: author},
	}

	return feed
}

func (fp *FeedService) getFeedUrl(author string) string {
	urlEncodedTitle := url.PathEscape(author)

	return fmt.Sprintf("/%s/%s/%s", fp.feedItemPath, urlEncodedTitle, defaultURLSuffix)
}

func (fp *FeedService) getFeedItemUrl(host string, parent_folder string, itemName string) string {
	urlEncodedItemName := url.PathEscape(itemName)
	// remove the suffix from the url
	urlPath := strings.TrimSuffix(fp.getFeedUrl(parent_folder), defaultURLSuffix)

	return fmt.Sprintf("%s%s%s", host, urlPath, urlEncodedItemName)
}
