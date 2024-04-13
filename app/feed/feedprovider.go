package feed

import (
	"reflect"
	"time"

	"github.com/gorilla/feeds"
	"github.com/jo-hoe/go-audio-rss-feeder/app/discovery"
)

type FeedProvider struct {
	audioSourceDirectory string
	feedTitle            string
	feedLink             string
	feedDescription      string
	feedAuthor           string
	feedCreated          time.Time
	feedImage            *feeds.Image
}

func NewFeedProvider(
	audioSourceDirectory string,
	feedTitle string,
	feedLink string,
	feedDescription string,
	feedAuthor string,
	feedCreated time.Time,
	feedCopyright string,
	feedImage *feeds.Image) *FeedProvider {
	return &FeedProvider{
		audioSourceDirectory: audioSourceDirectory,
		feedTitle:            feedTitle,
		feedLink:             feedLink,
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

	// create a feed parent and items for each audio file
	//for _, _ := range audioFiles {
	// checkout https://github.com/gorilla/feeds
	//}

	return nil, nil
}

func (fp *FeedProvider) createFeed() *feeds.Feed {
	now := time.Now()

	feed := &feeds.Feed{
		Title:       valueOrDefault(fp.feedTitle, "Rss Feed"),
		Link:        &feeds.Link{Href: valueOrDefault(fp.feedLink, "127.0.0.1:8080/rss.xml")},
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
