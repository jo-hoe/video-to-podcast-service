package feed

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jo-hoe/go-audio-rss-feeder/app/discovery"
	"github.com/jo-hoe/go-audio-rss-feeder/app/download"
	mp3joiner "github.com/jo-hoe/mp3-joiner"
)

type FeedProvider struct {
	audioSourceDirectory string
	feedBaseUrl          string
	feedTitle            string
	feedDescription      string
	feedAuthor           string
	feedCreated          time.Time
	feedImage            *feeds.Image
}

func NewFeedProvider(
	audioSourceDirectory string,
	feedBaseUrl string) *FeedProvider {
	return &FeedProvider{
		audioSourceDirectory: audioSourceDirectory,
		feedBaseUrl:          feedBaseUrl,
	}
}

func (fp *FeedProvider) GetFeed() (*feeds.RssFeed, error) {
	audioFilePaths, err := discovery.GetAudioFiles(fp.audioSourceDirectory)
	if err != nil {
		return nil, err
	}

	feed := fp.createFeed()
	for _, audioFilePath := range audioFilePaths {
		item, err := fp.createFeedItem(audioFilePath)
		if err != nil {
			return nil, err
		}

		feed.Items = append(feed.Items, item)
	}
	if len(feed.Items) > 0 {
		metadata, err := mp3joiner.GetFFmpegMetadataTag(audioFilePaths[0])
		if err != nil {
			return nil, err
		}
		feed.Image = &feeds.Image{
			Url: metadata[download.ThumbnailUrlTag],
		}
	}

	rssFeed := (&feeds.Rss{Feed: feed}).RssFeed()
	return rssFeed, nil
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
		Title:       valueOrDefault(audioMetadata["Title"], fileNameWithoutExtension),
		Link:        &feeds.Link{Href: fp.feedBaseUrl + fileInfo.Name()},
		Description: valueOrDefault(audioMetadata["Comment"], ""),
		Author:      &feeds.Author{Name: valueOrDefault(audioMetadata["Artist"], "")},
		Created:     fileInfo.ModTime(),
	}, nil
}

func (fp *FeedProvider) createFeed() *feeds.Feed {
	feed := &feeds.Feed{
		Title:       valueOrDefault(fp.feedTitle, "Rss Feed"),
		Link:        &feeds.Link{Href: valueOrDefault(fp.feedBaseUrl, "127.0.0.1:8080/rss.xml")},
		Description: valueOrDefault(fp.feedDescription, ""),
		Author:      &feeds.Author{Name: valueOrDefault(fp.feedAuthor, "")},
		Created:     valueOrDefault(fp.feedCreated, time.Now()),
		Image:       valueOrDefault(fp.feedImage, nil),
	}

	return feed
}

func (fp *FeedProvider) setFeedTitle(title string) {
	fp.feedTitle = title
}

func (fp *FeedProvider) setFeedDescription(description string) {
	fp.feedDescription = description
}

func (fp *FeedProvider) setFeedAuthor(author string) {
	fp.feedAuthor = author
}

func (fp *FeedProvider) setFeedCreationTime(creationTime time.Time) {
	fp.feedCreated = creationTime
}

func (fp *FeedProvider) setFeedImage(image *feeds.Image) {
	fp.feedImage = image
}

func valueOrDefault[T any](value, defaultValue T) T {
	reflectedValue := reflect.ValueOf(value)
	if reflectedValue.Kind() == reflect.Invalid {
		return defaultValue
	} else if reflectedValue.Kind() == reflect.String && reflectedValue.String() == "" {
		return defaultValue
	}
	return value
}
