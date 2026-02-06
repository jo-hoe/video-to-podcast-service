package database

import "database/sql"

type MockDatabase struct {
	Items                       map[string]*PodcastItem
	CreatePodcastItemFunc       func(item *PodcastItem) error
	GetPodcastItemByIDFunc      func(id string) (*PodcastItem, error)
	GetAllPodcastItemsFunc      func() ([]*PodcastItem, error)
	UpdatePodcastItemFunc       func(item *PodcastItem) error
	DeletePodcastItemFunc       func(id string) error
	GetPodcastItemsByAuthorFunc func(author string) ([]*PodcastItem, error)
}

func NewMockDatabase() *MockDatabase {
	return &MockDatabase{Items: make(map[string]*PodcastItem)}
}

func (m *MockDatabase) InsertReplacePodcastItem(item *PodcastItem) error {
	if m.CreatePodcastItemFunc != nil {
		return m.CreatePodcastItemFunc(item)
	}
	m.Items[item.ID] = item
	return nil
}

func (m *MockDatabase) GetPodcastItemByID(id string) (*PodcastItem, error) {
	if m.GetPodcastItemByIDFunc != nil {
		return m.GetPodcastItemByIDFunc(id)
	}
	item, ok := m.Items[id]
	if !ok {
		return nil, nil
	}
	return item, nil
}

func (m *MockDatabase) GetAllPodcastItems() ([]*PodcastItem, error) {
	if m.GetAllPodcastItemsFunc != nil {
		return m.GetAllPodcastItemsFunc()
	}
	var items []*PodcastItem
	for _, item := range m.Items {
		items = append(items, item)
	}
	return items, nil
}

func (m *MockDatabase) UpdatePodcastItem(item *PodcastItem) error {
	if m.UpdatePodcastItemFunc != nil {
		return m.UpdatePodcastItemFunc(item)
	}
	m.Items[item.ID] = item
	return nil
}

func (m *MockDatabase) DeletePodcastItem(id string) error {
	if m.DeletePodcastItemFunc != nil {
		return m.DeletePodcastItemFunc(id)
	}
	delete(m.Items, id)
	return nil
}

func (m *MockDatabase) GetPodcastItemsByAuthor(author string) ([]*PodcastItem, error) {
	if m.GetPodcastItemsByAuthorFunc != nil {
		return m.GetPodcastItemsByAuthorFunc(author)
	}
	var items []*PodcastItem
	for _, item := range m.Items {
		if item.Author == author {
			items = append(items, item)
		}
	}
	return items, nil
}

func (m *MockDatabase) InitializeDatabase() (*sql.DB, error) {
	return nil, nil
}

func (m *MockDatabase) CreateDatabase() (*sql.DB, error) {
	return nil, nil
}

func (m *MockDatabase) DoesDatabaseExist() bool {
	return true
}
