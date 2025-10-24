package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/yourname/turbo-db/internal/models"
)

var (
	apiURL string
	apiKey string
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "turbo-cli",
		Short: "Turbo DB CLI - Manage your databases",
		Long:  "A command-line interface for managing Turbo DB databases",
	}

	// Global flags
	rootCmd.PersistentFlags().StringVar(&apiURL, "url", "http://localhost:8080", "Turbo DB API URL")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", os.Getenv("TURBO_API_KEY"), "API key for authentication")

	// Add commands
	rootCmd.AddCommand(createCmd())
	rootCmd.AddCommand(listCmd())
	rootCmd.AddCommand(getCmd())
	rootCmd.AddCommand(deleteCmd())
	rootCmd.AddCommand(stopCmd())
	rootCmd.AddCommand(connectionCmd())

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func createCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			req := models.CreateDatabaseRequest{
				Name: name,
			}

			var resp models.CreateDatabaseResponse
			if err := apiRequest("POST", "/api/v1/databases", req, &resp); err != nil {
				return err
			}

			fmt.Printf("✓ Database created successfully!\n\n")
			fmt.Printf("ID:       %s\n", resp.Database.ID)
			fmt.Printf("Name:     %s\n", resp.Database.Name)
			fmt.Printf("Status:   %s\n", resp.Database.Status)
			fmt.Printf("HTTP URL: %s\n", resp.Database.HTTPUrl)
			fmt.Printf("Port:     %d\n", resp.Database.Port)

			return nil
		},
	}
}

func listCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all databases",
		RunE: func(cmd *cobra.Command, args []string) error {
			var resp models.ListDatabasesResponse
			if err := apiRequest("GET", "/api/v1/databases", nil, &resp); err != nil {
				return err
			}

			if resp.Count == 0 {
				fmt.Println("No databases found")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tPORT\tHTTP URL\tCREATED")
			fmt.Fprintln(w, "──\t────\t──────\t────\t────────\t───────")

			for _, db := range resp.Databases {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\t%s\n",
					truncate(db.ID, 8),
					db.Name,
					db.Status,
					db.Port,
					db.HTTPUrl,
					db.CreatedAt.Format("2006-01-02 15:04"),
				)
			}

			w.Flush()
			fmt.Printf("\nTotal: %d databases\n", resp.Count)

			return nil
		},
	}
}

func getCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get [id]",
		Short: "Get database details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			var db models.Database
			if err := apiRequest("GET", fmt.Sprintf("/api/v1/databases/%s", id), nil, &db); err != nil {
				return err
			}

			fmt.Printf("Database Details:\n\n")
			fmt.Printf("ID:         %s\n", db.ID)
			fmt.Printf("Name:       %s\n", db.Name)
			fmt.Printf("Status:     %s\n", db.Status)
			fmt.Printf("Path:       %s\n", db.Path)
			fmt.Printf("Port:       %d\n", db.Port)
			fmt.Printf("HTTP URL:   %s\n", db.HTTPUrl)
			fmt.Printf("Process ID: %d\n", db.ProcessID)
			fmt.Printf("Created:    %s\n", db.CreatedAt.Format(time.RFC3339))
			fmt.Printf("Updated:    %s\n", db.UpdatedAt.Format(time.RFC3339))

			return nil
		},
	}
}

func deleteCmd() *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete [id]",
		Short: "Delete a database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			if !confirm {
				fmt.Printf("Are you sure you want to delete database %s? This action cannot be undone.\n", id)
				fmt.Print("Type 'yes' to confirm: ")
				var input string
				fmt.Scanln(&input)
				if input != "yes" {
					fmt.Println("Deletion cancelled")
					return nil
				}
			}

			var resp map[string]interface{}
			if err := apiRequest("DELETE", fmt.Sprintf("/api/v1/databases/%s", id), nil, &resp); err != nil {
				return err
			}

			fmt.Printf("✓ Database deleted successfully\n")

			return nil
		},
	}

	cmd.Flags().BoolVarP(&confirm, "yes", "y", false, "Skip confirmation prompt")

	return cmd
}

func stopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop [id]",
		Short: "Stop a running database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			var resp map[string]interface{}
			if err := apiRequest("POST", fmt.Sprintf("/api/v1/databases/%s/stop", id), nil, &resp); err != nil {
				return err
			}

			fmt.Printf("✓ Database stopped successfully\n")

			return nil
		},
	}
}

func connectionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "connection [id]",
		Short: "Get database connection information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]

			var conn models.DatabaseConnectionInfo
			if err := apiRequest("GET", fmt.Sprintf("/api/v1/databases/%s/connection", id), nil, &conn); err != nil {
				return err
			}

			fmt.Printf("Connection Information:\n\n")
			fmt.Printf("Database ID: %s\n", conn.DatabaseID)
			fmt.Printf("Name:        %s\n", conn.Name)
			fmt.Printf("HTTP URL:    %s\n", conn.HTTPUrl)
			fmt.Printf("WS URL:      %s\n", conn.WSUrl)
			fmt.Printf("Status:      %s\n", conn.Status)

			return nil
		},
	}
}

// apiRequest makes an HTTP request to the API
func apiRequest(method, path string, body interface{}, result interface{}) error {
	url := apiURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-Key", apiKey)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		var errResp models.ErrorResponse
		if err := json.Unmarshal(respBody, &errResp); err == nil {
			return fmt.Errorf("API error: %s - %s", errResp.Error, errResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}
	}

	return nil
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max]
}
