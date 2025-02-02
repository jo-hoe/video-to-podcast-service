package feed

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jo-hoe/video-to-podcast-service/app/common"
	"github.com/jo-hoe/video-to-podcast-service/app/download"
	"github.com/jo-hoe/video-to-podcast-service/app/filemanagement"

	"github.com/gorilla/feeds"
	mp3joiner "github.com/jo-hoe/mp3-joiner"
)

const (
	defaultURLSuffix   = "rss.xml"
	defaultDescription = "Podcast Feed of"
	mp3KeyAttribute    = "artist"
)

type FeedService struct {
	audioSourceDirectory string
	feedBaseUrl          string
	feedBasePort         string
	feedItemPath         string
}

func NewFeedService(
	audioSourceDirectory string,
	feedBaseUrl string,
	feedBasePort string,
	feedItemPath string) *FeedService {
	return &FeedService{
		audioSourceDirectory: audioSourceDirectory,
		feedBaseUrl:          feedBaseUrl,
		feedBasePort:         feedBasePort,
		feedItemPath:         feedItemPath,
	}
}

func (fp *FeedService) GetFeeds() ([]*feeds.RssFeed, error) {
	feedCollector := make([]*feeds.Feed, 0)
	audioFilePaths, err := filemanagement.GetAudioFiles(fp.audioSourceDirectory)
	if err != nil {
		return nil, err
	}

	allItems := fp.createFeed("default")
	for _, audioFilePath := range audioFilePaths {
		directoryPath := filepath.Dir(audioFilePath)
		directoryName := filepath.Base(directoryPath)

		// either returns already created feed or nil
		feed := fp.getFeedWithAuthor(directoryName, feedCollector)
		if feed == nil {
			feed = fp.createFeed(directoryName)
			feedCollector = append(feedCollector, feed)
		}

		item, err := fp.createFeedItem(audioFilePath)
		if err != nil {
			return nil, err
		}

		metadata, err := mp3joiner.GetFFmpegMetadataTag(audioFilePath)
		if err != nil {
			return nil, err
		}
		feed.Items = append(feed.Items, item)
		if feed.Image == nil {
			feed.Image = &feeds.Image{
				Url:  metadata[download.ThumbnailUrlTag],
				Link: metadata[download.ThumbnailUrlTag],
			}
		}
		allItems.Items = append(allItems.Items, item)
	}

	results := make([]*feeds.RssFeed, 0)
	for _, item := range feedCollector {
		results = append(results, (&feeds.Rss{Feed: item}).RssFeed())
	}

	if len(allItems.Items) > 0 {
		results = append(results, (&feeds.Rss{Feed: allItems}).RssFeed())
	}

	return results, nil
}

func (fp *FeedService) getFeedWithAuthor(author string, feeds []*feeds.Feed) *feeds.Feed {
	for _, feed := range feeds {
		if feed.Author.Name == author {
			return feed
		}
	}
	return nil
}

func (fp *FeedService) createFeedItem(audioFilePath string) (*feeds.Item, error) {
	audioMetadata, err := mp3joiner.GetFFmpegMetadataTag(audioFilePath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(audioFilePath)
	if err != nil {
		return nil, err
	}
	fileNameWithoutExtension := strings.TrimSuffix(fileInfo.Name(), filepath.Ext(fileInfo.Name()))
	description := common.ValueOrDefault(audioMetadata[download.PodcastDescriptionTag], "")
	description = strings.ReplaceAll(description, "`n", "<br>")
	description = strings.ReplaceAll(description, "`r", "")

	uploadTime, err := time.Parse("20060102", audioMetadata[download.DateTag])
	if err != nil {
		log.Printf("could not parse date tag, reverting to default. error: %v", err)
		uploadTime = fileInfo.ModTime()
	}

	return &feeds.Item{
		Title:       common.ValueOrDefault(audioMetadata["Title"], fileNameWithoutExtension),
		Link:        &feeds.Link{Href: fp.getFeedItemUrl(audioMetadata[mp3KeyAttribute], fileInfo.Name())},
		Description: description,
		Author:      &feeds.Author{Name: common.ValueOrDefault(audioMetadata[mp3KeyAttribute], "")},
		Created:     uploadTime,
	}, nil
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

	return fmt.Sprintf("%s:%s/%s/%s/%s", fp.feedBaseUrl, fp.feedBasePort, fp.feedItemPath, urlEncodedTitle, defaultURLSuffix)
}

func (fp *FeedService) getFeedItemUrl(author string, itemName string) string {
	urlEncodedItemName := url.PathEscape(itemName)
	// remove the suffix from the url
	urlPath := strings.TrimSuffix(fp.getFeedUrl(author), fmt.Sprintf("/%s", defaultURLSuffix))

	return fmt.Sprintf("%s/%s", urlPath, urlEncodedItemName)
}
