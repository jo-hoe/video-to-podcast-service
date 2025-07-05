package database

type Database interface {
	CreatePodcastItem(item *PodcastItem) error
	GetPodcastItemByID(id string) (*PodcastItem, error)
	GetAllPodcastItems() ([]*PodcastItem, error)
	UpdatePodcastItem(item *PodcastItem) error
	DeletePodcastItem(id string) error
	GetPodcastItemsByAuthor(author string) ([]*PodcastItem, error)
}
