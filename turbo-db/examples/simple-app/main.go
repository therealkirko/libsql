package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// TurboDBClient is a simple client for Turbo DB API
type TurboDBClient struct {
	BaseURL string
	APIKey  string
	Client  *http.Client
}

// Database represents a database in Turbo DB
type Database struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	Port      int       `json:"port"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	HTTPUrl   string    `json:"http_url"`
}

// CreateDatabaseRequest is the request to create a database
type CreateDatabaseRequest struct {
	Name string `json:"name"`
}

// CreateDatabaseResponse is the response after creating a database
type CreateDatabaseResponse struct {
	Database Database `json:"database"`
	Message  string   `json:"message"`
}

// ListDatabasesResponse is the response for listing databases
type ListDatabasesResponse struct {
	Databases []Database `json:"databases"`
	Count     int        `json:"count"`
}

// NewClient creates a new Turbo DB client
func NewClient(baseURL, apiKey string) *TurboDBClient {
	return &TurboDBClient{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// CreateDatabase creates a new database
func (c *TurboDBClient) CreateDatabase(name string) (*Database, error) {
	req := CreateDatabaseRequest{Name: name}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequest("POST", c.BaseURL+"/api/v1/databases", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	if c.APIKey != "" {
		httpReq.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result CreateDatabaseResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return &result.Database, nil
}

// ListDatabases lists all databases
func (c *TurboDBClient) ListDatabases() ([]Database, error) {
	httpReq, err := http.NewRequest("GET", c.BaseURL+"/api/v1/databases", nil)
	if err != nil {
		return nil, err
	}

	if c.APIKey != "" {
		httpReq.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result ListDatabasesResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	return result.Databases, nil
}

// GetDatabase gets a specific database
func (c *TurboDBClient) GetDatabase(id string) (*Database, error) {
	httpReq, err := http.NewRequest("GET", c.BaseURL+"/api/v1/databases/"+id, nil)
	if err != nil {
		return nil, err
	}

	if c.APIKey != "" {
		httpReq.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var db Database
	if err := json.Unmarshal(respBody, &db); err != nil {
		return nil, err
	}

	return &db, nil
}

// DeleteDatabase deletes a database
func (c *TurboDBClient) DeleteDatabase(id string) error {
	httpReq, err := http.NewRequest("DELETE", c.BaseURL+"/api/v1/databases/"+id, nil)
	if err != nil {
		return err
	}

	if c.APIKey != "" {
		httpReq.Header.Set("X-API-Key", c.APIKey)
	}

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	return nil
}

func main() {
	// Create a Turbo DB client
	client := NewClient("http://localhost:8080", "")

	fmt.Println("=== Turbo DB Client Example ===\n")

	// Create a database
	fmt.Println("1. Creating a new database...")
	db, err := client.CreateDatabase("example-app-db")
	if err != nil {
		log.Fatalf("Failed to create database: %v", err)
	}
	fmt.Printf("✓ Created database: %s (ID: %s)\n", db.Name, db.ID)
	fmt.Printf("  HTTP URL: %s\n", db.HTTPUrl)
	fmt.Printf("  Port: %d\n", db.Port)
	fmt.Printf("  Status: %s\n\n", db.Status)

	// List all databases
	fmt.Println("2. Listing all databases...")
	databases, err := client.ListDatabases()
	if err != nil {
		log.Fatalf("Failed to list databases: %v", err)
	}
	fmt.Printf("✓ Found %d database(s):\n", len(databases))
	for i, d := range databases {
		fmt.Printf("  %d. %s (Port: %d, Status: %s)\n", i+1, d.Name, d.Port, d.Status)
	}
	fmt.Println()

	// Get database details
	fmt.Println("3. Getting database details...")
	dbDetails, err := client.GetDatabase(db.ID)
	if err != nil {
		log.Fatalf("Failed to get database: %v", err)
	}
	fmt.Printf("✓ Database details:\n")
	fmt.Printf("  Name: %s\n", dbDetails.Name)
	fmt.Printf("  ID: %s\n", dbDetails.ID)
	fmt.Printf("  Path: %s\n", dbDetails.Path)
	fmt.Printf("  Port: %d\n", dbDetails.Port)
	fmt.Printf("  Status: %s\n", dbDetails.Status)
	fmt.Printf("  Created: %s\n\n", dbDetails.CreatedAt.Format(time.RFC3339))

	// At this point, you can connect to the database using the HTTP URL
	fmt.Println("4. Connecting to the database...")
	fmt.Printf("You can now connect to the database at: %s\n", dbDetails.HTTPUrl)
	fmt.Println("\nExample using libsql-client-go:")
	fmt.Printf(`
	import "github.com/tursodatabase/libsql-client-go/libsql"

	db, err := libsql.NewEmbeddedReplicaDatabase("%s", "")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Execute queries
	conn, err := db.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	_, err = conn.Execute("CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT)")
	if err != nil {
		log.Fatal(err)
	}
`, dbDetails.HTTPUrl)

	// Wait a bit
	fmt.Println("\n5. Waiting for 5 seconds before cleanup...")
	time.Sleep(5 * time.Second)

	// Clean up - delete the database
	fmt.Println("6. Cleaning up - deleting the database...")
	if err := client.DeleteDatabase(db.ID); err != nil {
		log.Fatalf("Failed to delete database: %v", err)
	}
	fmt.Printf("✓ Deleted database: %s\n", db.Name)

	fmt.Println("\n=== Example Complete ===")
}
