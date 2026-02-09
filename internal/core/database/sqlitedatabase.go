package database

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultDatabaseName     = "podcast_items"
	defaultDatabaseExt      = ".db"
	defaultDatabaseFileName = defaultDatabaseName + defaultDatabaseExt
)

// SQLiteDatabase implements the Database interface using SQLite and prepared statements.
type SQLiteDatabase struct {
	db               *sql.DB
	connectionString string
}

func NewSQLiteDatabase(connectionString string) *SQLiteDatabase {
	return &SQLiteDatabase{
		connectionString: connectionString,
	}
}

func (s *SQLiteDatabase) InitializeDatabase() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", s.connectionString)
	if err != nil {
		return nil, err
	}
	// Optionally, check if the connection is valid
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	s.db = db
	return db, nil
}

func (s *SQLiteDatabase) CreateDatabase() (*sql.DB, error) {
	// Handle empty connection string by setting a default file-based database
	if s.connectionString == "" {
		s.connectionString = defaultDatabaseFileName
	}
	db, err := sql.Open("sqlite3", s.connectionString)
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
		created_at DATETIME,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`, defaultDatabaseName)
	_, err = db.Exec(createTableStmt)
	if err != nil {
		_ = db.Close()
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

func (s *SQLiteDatabase) InsertReplacePodcastItem(item *PodcastItem) error {
	stmt, err := s.db.Prepare(fmt.Sprintf(`INSERT OR REPLACE INTO %s (
		id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at, updated_at
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, defaultDatabaseName))
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()
	_, err = stmt.Exec(
		item.ID, item.Title, item.Description, item.Author, item.Thumbnail,
		item.DurationInMilliseconds, item.VideoURL, item.AudioFilePath, item.CreatedAt.UTC(), item.UpdatedAt.UTC(),
	)
	return err
}

func (s *SQLiteDatabase) GetAllPodcastItems() ([]*PodcastItem, error) {
	rows, err := s.db.Query(fmt.Sprintf(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at, updated_at FROM %s`, defaultDatabaseName))
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var items []*PodcastItem
		for rows.Next() {
			item := &PodcastItem{}
			err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.Author, &item.Thumbnail, &item.DurationInMilliseconds, &item.VideoURL, &item.AudioFilePath, &item.CreatedAt, &item.UpdatedAt)
			if err != nil {
				return nil, err
			}
			item.CreatedAt = item.CreatedAt.UTC()
			item.UpdatedAt = item.UpdatedAt.UTC()
			items = append(items, item)
		}
	return items, nil
}

func (m *SQLiteDatabase) DeletePodcastItem(id string) error {
	stmt, err := m.db.Prepare(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, defaultDatabaseName))
	if err != nil {
		return err
	}
	defer func() { _ = stmt.Close() }()
	_, err = stmt.Exec(id)
	if err != nil {
		return fmt.Errorf("failed to delete podcast item with id %s: %w", id, err)
	}
	return nil
}

func (s *SQLiteDatabase) GetPodcastItemByID(id string) (*PodcastItem, error) {
	stmt, err := s.db.Prepare(fmt.Sprintf(`SELECT id, title, description, author, thumbnail, duration_in_milliseconds, video_url, audio_file_path, created_at, updated_at FROM %s WHERE id = ?`, defaultDatabaseName))
	if err != nil {
		return nil, err
	}
	defer func() { _ = stmt.Close() }()
	item := &PodcastItem{}
	err = stmt.QueryRow(id).Scan(&item.ID, &item.Title, &item.Description, &item.Author, &item.Thumbnail, &item.DurationInMilliseconds, &item.VideoURL, &item.AudioFilePath, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("podcast item with id %s not found", id)
		}
		return nil, err
	}
	item.CreatedAt = item.CreatedAt.UTC()
	item.UpdatedAt = item.UpdatedAt.UTC()
	return item, nil
}

// Close closes the database connection.
func (s *SQLiteDatabase) CloseConnection() error {
	return s.db.Close()
}

func (s *SQLiteDatabase) DropDatabase() error {
	if s.db == nil {
		return fmt.Errorf("database connection is not initialized")
	}
	// Close the database connection before dropping it
	if err := s.CloseConnection(); err != nil {
		return err
	}
	// Remove the database file
	return os.Remove(s.connectionString)
}
