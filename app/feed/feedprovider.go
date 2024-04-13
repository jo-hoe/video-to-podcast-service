package feed

import (
	"reflect"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jo-hoe/go-audio-rss-feeder/app/discovery"
)

type FeedProvider struct {
	audioSourceDirectory string
	feedBaseUrl          string
	feedTitle            string
	feedLink             string
	feedDescription      string
	feedAuthor           string
	feedCreated          time.Time
	feedImage            *feeds.Image
}

func NewFeedProvider(
	audioSourceDirectory string,
	feedBaseUrl string,
	feedTitle string,
	feedDescription string,
	feedAuthor string,
	feedCreated time.Time,
	feedImage *feeds.Image) *FeedProvider {
	return &FeedProvider{
		audioSourceDirectory: audioSourceDirectory,
		feedBaseUrl:          feedBaseUrl,
		feedTitle:            feedTitle,
		feedDescription:      feedDescription,
		feedAuthor:           feedAuthor,
		feedCreated:          feedCreated,
		feedImage:            feedImage,
	}
}

func (fp *FeedProvider) GetFeed() (*feeds.RssFeed, error) {
	_, err := discovery.GetAudioFiles(fp.audioSourceDirectory)
	if err != nil {
		return nil, err
	}

	feed := fp.createFeed()

	rssFeed := (&feeds.Rss{Feed: feed}).RssFeed()
	return rssFeed, nil
}

func (fp *FeedProvider) createFeed() *feeds.Feed {
	now := time.Now()

	feed := &feeds.Feed{
		Title:       valueOrDefault(fp.feedTitle, "Rss Feed"),
		Link:        &feeds.Link{Href: valueOrDefault(fp.feedBaseUrl, "127.0.0.1:8080/rss.xml")},
		Description: valueOrDefault(fp.feedDescription, ""),
		Author:      &feeds.Author{Name: valueOrDefault(fp.feedAuthor, "")},
		Created:     now,
		Image:       valueOrDefault(fp.feedImage, nil),
	}

	return feed
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
