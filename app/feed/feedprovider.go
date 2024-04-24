package feed

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/feeds"
	"github.com/jo-hoe/go-audio-rss-feeder/app/common"
	"github.com/jo-hoe/go-audio-rss-feeder/app/discovery"
	"github.com/jo-hoe/go-audio-rss-feeder/app/download"
	mp3joiner "github.com/jo-hoe/mp3-joiner"
)

const (
	defaultURL         = "127.0.0.1:8080"
	defaultURLSuffix   = "rss.xml"
	defaultTitlePrefix = "Podcast Feed of"
	defaultDescription = defaultTitlePrefix
	mp3KeyAttribute    = "artist"
)

type FeedProvider struct {
	audioSourceDirectory string
	feedBaseUrl          string
}

func NewFeedProvider(
	audioSourceDirectory string,
	feedBaseUrl string) *FeedProvider {
	return &FeedProvider{
		audioSourceDirectory: audioSourceDirectory,
		feedBaseUrl:          feedBaseUrl,
	}
}

func (fp *FeedProvider) GetFeeds() ([]*feeds.RssFeed, error) {
	feedCollector := make([]*feeds.Feed, 0)
	audioFilePaths, err := discovery.GetAudioFiles(fp.audioSourceDirectory)
	if err != nil {
		return nil, err
	}

	for _, audioFilePath := range audioFilePaths {
		metadata, err := mp3joiner.GetFFmpegMetadataTag(audioFilePath)
		if err != nil {
			return nil, err
		}

		if metadata[mp3KeyAttribute] == "" {
			log.Printf("no '%s' found for file '%s' - skipping file", mp3KeyAttribute, filepath.Base(audioFilePath))
			continue
		}

		// either returns already created feed or nil
		feed := fp.getFeedWithAuthor(metadata[mp3KeyAttribute], feedCollector)
		if feed == nil {
			feed = fp.createFeed(metadata[mp3KeyAttribute])
			feedCollector = append(feedCollector, feed)
		}

		item, err := fp.createFeedItem(audioFilePath)
		if err != nil {
			return nil, err
		}

		feed.Items = append(feed.Items, item)
		if feed.Image == nil {
			feed.Image = &feeds.Image{
				Url: metadata[download.ThumbnailUrlTag],
			}
		}
	}

	results := make([]*feeds.RssFeed, 0)
	for _, item := range feedCollector {
		results = append(results, (&feeds.Rss{Feed: item}).RssFeed())
	}

	return results, nil
}

func (fp *FeedProvider) getFeedWithAuthor(author string, feeds []*feeds.Feed) *feeds.Feed {
	for _, feed := range feeds {
		if feed.Author.Name == author {
			return feed
		}
	}
	return nil
}

func (fp *FeedProvider) createFeedItem(audioFilePath string) (*feeds.Item, error) {
	audioMetadata, err := mp3joiner.GetFFmpegMetadataTag(audioFilePath)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(audioFilePath)
	if err != nil {
		return nil, err
	}
	fileNameWithoutExtension := strings.TrimSuffix(fileInfo.Name(), filepath.Ext(fileInfo.Name()))

	return &feeds.Item{
		Title:       common.ValueOrDefault(audioMetadata["Title"], fileNameWithoutExtension),
		Link:        &feeds.Link{Href: fp.getFeedItemUrl(audioMetadata[mp3KeyAttribute], fileInfo.Name())},
		Description: common.ValueOrDefault(audioMetadata["Comment"], ""),
		Author:      &feeds.Author{Name: common.ValueOrDefault(audioMetadata[mp3KeyAttribute], "")},
		Created:     fileInfo.ModTime(),
	}, nil
}

func (fp *FeedProvider) createFeed(author string) *feeds.Feed {
	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s %s", defaultTitlePrefix, author),
		Link:        &feeds.Link{Href: fp.getFeedUrl(author)},
		Description: fmt.Sprintf("%s %s", defaultDescription, author),
		Author:      &feeds.Author{Name: author},
	}

	return feed
}

func (fp *FeedProvider) getFeedUrl(author string) string {
	urlEncodedTitle := url.PathEscape(author)

	return fmt.Sprintf("%s/%s/%s", common.ValueOrDefault(fp.feedBaseUrl, defaultURL), urlEncodedTitle, defaultURLSuffix)
}

func (fp *FeedProvider) getFeedItemUrl(author string, itemName string) string {
	urlEncodedItemName := url.PathEscape(itemName)

	return fmt.Sprintf("%s/%s", fp.getFeedUrl(author), urlEncodedItemName)
}
