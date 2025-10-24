package models

import (
	"time"
)

// Database represents a managed database instance
type Database struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Path        string    `json:"path" db:"path"`
	Port        int       `json:"port" db:"port"`
	Status      string    `json:"status" db:"status"` // running, stopped, error
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	ProcessID   int       `json:"process_id,omitempty" db:"process_id"`
	HTTPUrl     string    `json:"http_url" db:"http_url"`
}

// CreateDatabaseRequest represents the request to create a new database
type CreateDatabaseRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateDatabaseResponse represents the response after creating a database
type CreateDatabaseResponse struct {
	Database    Database `json:"database"`
	Message     string   `json:"message"`
}

// ListDatabasesResponse represents the response for listing databases
type ListDatabasesResponse struct {
	Databases []Database `json:"databases"`
	Count     int        `json:"count"`
}

// DatabaseConnectionInfo provides connection details for a database
type DatabaseConnectionInfo struct {
	DatabaseID  string `json:"database_id"`
	Name        string `json:"name"`
	HTTPUrl     string `json:"http_url"`
	WSUrl       string `json:"ws_url"`
	Status      string `json:"status"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}
