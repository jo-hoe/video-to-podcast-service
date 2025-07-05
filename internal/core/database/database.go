package database

import "database/sql"

type Database interface {
	InitializeDatabase(connectionsString string) (*sql.DB, error)
	CreateDatabase(connectionsString string) (*sql.DB, error)
	DoesDatabaseExist() bool

	CreatePodcastItem(item *PodcastItem) error
	GetPodcastItemByID(id string) (*PodcastItem, error)
	GetAllPodcastItems() ([]*PodcastItem, error)
	UpdatePodcastItem(item *PodcastItem) error
	DeletePodcastItem(id string) error
	GetPodcastItemsByAuthor(author string) ([]*PodcastItem, error)
}
