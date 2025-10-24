package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/yourname/turbo-db/internal/models"
)

// Store manages the metadata database
type Store struct {
	db *sql.DB
}

// NewStore creates a new Store instance
func NewStore(dbPath string) (*Store, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open metadata database: %w", err)
	}

	store := &Store{db: db}

	if err := store.initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize metadata database: %w", err)
	}

	return store, nil
}

// initialize creates the necessary tables
func (s *Store) initialize() error {
	schema := `
	CREATE TABLE IF NOT EXISTS databases (
		id TEXT PRIMARY KEY,
		name TEXT UNIQUE NOT NULL,
		path TEXT NOT NULL,
		port INTEGER NOT NULL,
		status TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL,
		process_id INTEGER DEFAULT 0,
		http_url TEXT NOT NULL
	);

	CREATE INDEX IF NOT EXISTS idx_databases_name ON databases(name);
	CREATE INDEX IF NOT EXISTS idx_databases_status ON databases(status);
	`

	_, err := s.db.Exec(schema)
	return err
}

// Create adds a new database record
func (s *Store) Create(db *models.Database) error {
	query := `
		INSERT INTO databases (id, name, path, port, status, created_at, updated_at, process_id, http_url)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		db.ID,
		db.Name,
		db.Path,
		db.Port,
		db.Status,
		db.CreatedAt,
		db.UpdatedAt,
		db.ProcessID,
		db.HTTPUrl,
	)

	if err != nil {
		return fmt.Errorf("failed to create database record: %w", err)
	}

	return nil
}

// Get retrieves a database by ID
func (s *Store) Get(id string) (*models.Database, error) {
	query := `
		SELECT id, name, path, port, status, created_at, updated_at, process_id, http_url
		FROM databases WHERE id = ?
	`

	var db models.Database
	err := s.db.QueryRow(query, id).Scan(
		&db.ID,
		&db.Name,
		&db.Path,
		&db.Port,
		&db.Status,
		&db.CreatedAt,
		&db.UpdatedAt,
		&db.ProcessID,
		&db.HTTPUrl,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("database not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	return &db, nil
}

// GetByName retrieves a database by name
func (s *Store) GetByName(name string) (*models.Database, error) {
	query := `
		SELECT id, name, path, port, status, created_at, updated_at, process_id, http_url
		FROM databases WHERE name = ?
	`

	var db models.Database
	err := s.db.QueryRow(query, name).Scan(
		&db.ID,
		&db.Name,
		&db.Path,
		&db.Port,
		&db.Status,
		&db.CreatedAt,
		&db.UpdatedAt,
		&db.ProcessID,
		&db.HTTPUrl,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("database not found")
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get database: %w", err)
	}

	return &db, nil
}

// List retrieves all databases
func (s *Store) List() ([]models.Database, error) {
	query := `
		SELECT id, name, path, port, status, created_at, updated_at, process_id, http_url
		FROM databases ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %w", err)
	}
	defer rows.Close()

	var databases []models.Database
	for rows.Next() {
		var db models.Database
		err := rows.Scan(
			&db.ID,
			&db.Name,
			&db.Path,
			&db.Port,
			&db.Status,
			&db.CreatedAt,
			&db.UpdatedAt,
			&db.ProcessID,
			&db.HTTPUrl,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan database row: %w", err)
		}
		databases = append(databases, db)
	}

	return databases, nil
}

// Update updates a database record
func (s *Store) Update(db *models.Database) error {
	query := `
		UPDATE databases
		SET status = ?, updated_at = ?, process_id = ?
		WHERE id = ?
	`

	db.UpdatedAt = time.Now()

	_, err := s.db.Exec(query, db.Status, db.UpdatedAt, db.ProcessID, db.ID)
	if err != nil {
		return fmt.Errorf("failed to update database: %w", err)
	}

	return nil
}

// Delete removes a database record
func (s *Store) Delete(id string) error {
	query := `DELETE FROM databases WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete database: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("database not found")
	}

	return nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}
