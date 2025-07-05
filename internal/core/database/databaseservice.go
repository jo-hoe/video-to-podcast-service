package database

import "database/sql"

type DatabaseService interface {
	InitializeDatabase() (*sql.DB, error)
	CreateDatabase() (*sql.DB, error)
	DoesDatabaseExist() bool

	CreatePodcastItem(item *PodcastItem) error // CreatePodcastItem creates a new podcast item in the database and overwrites it if it already exists.
	GetPodcastItemByID(id string) (*PodcastItem, error)
	GetAllPodcastItems() ([]*PodcastItem, error)
	UpdatePodcastItem(item *PodcastItem) error
	DeletePodcastItem(id string) error
	GetPodcastItemsByAuthor(author string) ([]*PodcastItem, error)
}
