package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultDatabaseName     = "podcast_items"
	defaultDatabaseExt      = ".db"
	defaultDatabaseFileName = defaultDatabaseName + defaultDatabaseExt
)

// SQLiteDatabase implements the Database interface using SQLite and prepared statements.
type SQLiteDatabase struct {
	db *sql.DB
}

// NewSQLiteDatabase creates a new SQLiteDatabase instance.
// Implements Database interface
func (s *SQLiteDatabase) InitializeDatabase(connectionString string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}
	// Optionally, check if the connection is valid
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	s.db = db
	return db, nil
}

func (s *SQLiteDatabase) CreateDatabase(connectionString string) (*sql.DB, error) {
	// Handle empty connection string by setting a default file-based database
	if connectionString == "" {
		connectionString = defaultDatabaseFileName
	}
	db, err := sql.Open("sqlite3", connectionString)
	if err != nil {
		return nil, err
	}
	createTableStmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id TEXT PRIMARY KEY,
		title TEXT,
		description TEXT,
		author TEXT,
		thumbnail TEXT,
		duration_in_milliseconds INTEGER,
		video_url TEXT,
		audio_file_path TEXT,
		created_at DATETIME
	)`, defaultDatabaseName)
	_, err = db.Exec(createTableStmt)
	if err != nil {
		db.Close()
		return nil, err
	}
	s.db = db
	return db, nil
}

func (s *SQLiteDatabase) DoesDatabaseExist() bool {
	// Check if the database file exists
	if s.db == nil {
		return false
	}
	var exists bool
	// Check if the table exists in the database
	err := s.db.QueryRow(fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type='table' AND name='%s')`, defaultDatabaseName)).Scan(&exists)
	if err != nil {
		return false
	}
	return exists
}

func (s *SQLiteDatabase) CreatePodcastItem(item *PodcastItem) error {
	stmt, err := s.db.Prepare(fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, defaultDatabaseName))
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
	stmt, err := s.db.Prepare(fmt.Sprintf(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at FROM %s WHERE id = ?`, defaultDatabaseName))
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
	rows, err := s.db.Query(fmt.Sprintf(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at FROM %s`, defaultDatabaseName))
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
	stmt, err := s.db.Prepare(fmt.Sprintf(`UPDATE %s SET title=?, description=?, author=?, thumbnail=?, duration_in_milliseconds=?, video_url=?, audio_file_path=?, created_at=? WHERE id=?`, defaultDatabaseName))
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
	stmt, err := s.db.Prepare(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, defaultDatabaseName))
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	return err
}

func (s *SQLiteDatabase) GetPodcastItemsByAuthor(author string) ([]*PodcastItem, error) {
	stmt, err := s.db.Prepare(fmt.Sprintf(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at FROM %s WHERE author = ?`, defaultDatabaseName))
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
