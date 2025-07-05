package database

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteDatabase implements the Database interface using SQLite and prepared statements.
type SQLiteDatabase struct {
	db *sql.DB
}

// NewSQLiteDatabase creates a new SQLiteDatabase instance.
func NewSQLiteDatabase(dataSourceName string) (*SQLiteDatabase, error) {
	db, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}
	return &SQLiteDatabase{db: db}, nil
}

func (s *SQLiteDatabase) CreatePodcastItem(item *PodcastItem) error {
	stmt, err := s.db.Prepare(`INSERT INTO podcast_items (
		id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		item.ID, item.Title, item.Description, item.Author, item.Thumbnail,
		item.DurationInMilliseconds, item.VideoURL, item.AudioFilePath, item.CreatedAt,
	)
	return err
}

func (s *SQLiteDatabase) GetPodcastItemByID(id string) (*PodcastItem, error) {
	stmt, err := s.db.Prepare(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at FROM podcast_items WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	item := &PodcastItem{}
	row := stmt.QueryRow(id)
	err = row.Scan(&item.ID, &item.Title, &item.Description, &item.Author, &item.Thumbnail, &item.DurationInMilliseconds, &item.VideoURL, &item.AudioFilePath, &item.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return item, nil
}

func (s *SQLiteDatabase) GetAllPodcastItems() ([]*PodcastItem, error) {
	rows, err := s.db.Query(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at FROM podcast_items`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PodcastItem
	for rows.Next() {
		item := &PodcastItem{}
		err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Author, &item.Thumbnail, &item.DurationInMilliseconds, &item.VideoURL, &item.AudioFilePath, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *SQLiteDatabase) UpdatePodcastItem(item *PodcastItem) error {
	stmt, err := s.db.Prepare(`UPDATE podcast_items SET title=?, description=?, author=?, thumbnail=?, duration_in_milliseconds=?, video_url=?, audio_file_path=?, created_at=? WHERE id=?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(
		item.Title, item.Description, item.Author, item.Thumbnail, item.DurationInMilliseconds, item.VideoURL, item.AudioFilePath, item.CreatedAt, item.ID,
	)
	return err
}

func (s *SQLiteDatabase) DeletePodcastItem(id string) error {
	stmt, err := s.db.Prepare(`DELETE FROM podcast_items WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}

func (s *SQLiteDatabase) GetPodcastItemsByAuthor(author string) ([]*PodcastItem, error) {
	stmt, err := s.db.Prepare(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at FROM podcast_items WHERE author = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()
	rows, err := stmt.Query(author)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []*PodcastItem
	for rows.Next() {
		item := &PodcastItem{}
		err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Author, &item.Thumbnail, &item.DurationInMilliseconds, &item.VideoURL, &item.AudioFilePath, &item.CreatedAt)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

// Close closes the database connection.
func (s *SQLiteDatabase) Close() error {
	return s.db.Close()
}
