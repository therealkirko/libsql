package manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/yourname/turbo-db/internal/config"
	"github.com/yourname/turbo-db/internal/database"
	"github.com/yourname/turbo-db/internal/models"
)

// Manager handles database lifecycle management
type Manager struct {
	config    *config.Config
	store     *database.Store
	processes map[string]*exec.Cmd
	portPool  *PortPool
	mu        sync.RWMutex
}

// NewManager creates a new database manager
func NewManager(cfg *config.Config, store *database.Store) (*Manager, error) {
	// Ensure databases directory exists
	if err := os.MkdirAll(cfg.DatabasesDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create databases directory: %w", err)
	}

	portPool := NewPortPool(cfg.PortRangeStart, cfg.PortRangeEnd)

	m := &Manager{
		config:    cfg,
		store:     store,
		processes: make(map[string]*exec.Cmd),
		portPool:  portPool,
	}

	// Restore running databases on startup
	if err := m.restoreDatabases(); err != nil {
		return nil, fmt.Errorf("failed to restore databases: %w", err)
	}

	return m, nil
}

// CreateDatabase creates and starts a new database instance
func (m *Manager) CreateDatabase(name string) (*models.Database, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if database already exists
	existing, _ := m.store.GetByName(name)
	if existing != nil {
		return nil, fmt.Errorf("database with name '%s' already exists", name)
	}

	// Allocate a port
	port, err := m.portPool.Allocate()
	if err != nil {
		return nil, fmt.Errorf("failed to allocate port: %w", err)
	}

	// Create database model
	db := &models.Database{
		ID:        uuid.New().String(),
		Name:      name,
		Path:      filepath.Join(m.config.DatabasesDir, name+".db"),
		Port:      port,
		Status:    "creating",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		HTTPUrl:   fmt.Sprintf("http://localhost:%d", port),
	}

	// Save to store
	if err := m.store.Create(db); err != nil {
		m.portPool.Release(port)
		return nil, fmt.Errorf("failed to save database metadata: %w", err)
	}

	// Start the libsql server process
	if err := m.startDatabase(db); err != nil {
		// Clean up on failure
		m.store.Delete(db.ID)
		m.portPool.Release(port)
		return nil, fmt.Errorf("failed to start database: %w", err)
	}

	db.Status = "running"
	m.store.Update(db)

	return db, nil
}

// startDatabase starts a libsql server process for the given database
func (m *Manager) startDatabase(db *models.Database) error {
	// Build the command to start sqld
	cmd := exec.Command(
		m.config.LibSQLBinary,
		"--http-listen-addr", fmt.Sprintf("127.0.0.1:%d", db.Port),
		"--db-path", db.Path,
	)

	// Set up logging (you can customize this)
	logDir := filepath.Join(m.config.DatabasesDir, "logs")
	os.MkdirAll(logDir, 0755)

	logFile, err := os.Create(filepath.Join(logDir, fmt.Sprintf("%s.log", db.Name)))
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	// Start the process
	if err := cmd.Start(); err != nil {
		logFile.Close()
		return fmt.Errorf("failed to start sqld process: %w", err)
	}

	// Store the process
	m.processes[db.ID] = cmd
	db.ProcessID = cmd.Process.Pid

	// Wait a bit to ensure the process started successfully
	time.Sleep(1 * time.Second)

	// Check if process is still running
	if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
		return fmt.Errorf("process failed to start properly")
	}

	// Start a goroutine to monitor the process
	go m.monitorProcess(db.ID, cmd)

	return nil
}

// monitorProcess monitors a database process and updates status
func (m *Manager) monitorProcess(dbID string, cmd *exec.Cmd) {
	err := cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.processes, dbID)

	db, getErr := m.store.Get(dbID)
	if getErr != nil {
		return
	}

	if err != nil {
		db.Status = "error"
	} else {
		db.Status = "stopped"
	}
	db.ProcessID = 0

	m.store.Update(db)
	m.portPool.Release(db.Port)
}

// StopDatabase stops a running database instance
func (m *Manager) StopDatabase(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	db, err := m.store.Get(id)
	if err != nil {
		return err
	}

	cmd, exists := m.processes[id]
	if !exists || cmd.Process == nil {
		return fmt.Errorf("database process not found")
	}

	// Send SIGTERM to gracefully shutdown
	if err := cmd.Process.Signal(syscall.SIGTERM); err != nil {
		// If SIGTERM fails, force kill
		cmd.Process.Kill()
	}

	delete(m.processes, id)
	db.Status = "stopped"
	db.ProcessID = 0
	m.store.Update(db)

	return nil
}

// DeleteDatabase stops and removes a database
func (m *Manager) DeleteDatabase(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	db, err := m.store.Get(id)
	if err != nil {
		return err
	}

	// Stop the process if running
	if cmd, exists := m.processes[id]; exists && cmd.Process != nil {
		cmd.Process.Kill()
		delete(m.processes, id)
	}

	// Remove database file
	if err := os.Remove(db.Path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove database file: %w", err)
	}

	// Remove WAL files if they exist
	os.Remove(db.Path + "-wal")
	os.Remove(db.Path + "-shm")

	// Remove from metadata store
	if err := m.store.Delete(id); err != nil {
		return fmt.Errorf("failed to remove database metadata: %w", err)
	}

	// Release the port
	m.portPool.Release(db.Port)

	return nil
}

// GetDatabase retrieves database information
func (m *Manager) GetDatabase(id string) (*models.Database, error) {
	return m.store.Get(id)
}

// ListDatabases lists all databases
func (m *Manager) ListDatabases() ([]models.Database, error) {
	return m.store.List()
}

// restoreDatabases restores databases that were running before shutdown
func (m *Manager) restoreDatabases() error {
	databases, err := m.store.List()
	if err != nil {
		return err
	}

	for _, db := range databases {
		// Mark ports as allocated
		m.portPool.MarkAllocated(db.Port)

		// Try to restart databases that were running
		if db.Status == "running" {
			if err := m.startDatabase(&db); err != nil {
				db.Status = "error"
				m.store.Update(&db)
			} else {
				db.Status = "running"
				m.store.Update(&db)
			}
		}
	}

	return nil
}

// Shutdown gracefully shuts down all databases
func (m *Manager) Shutdown() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, cmd := range m.processes {
		if cmd.Process != nil {
			cmd.Process.Signal(syscall.SIGTERM)

			db, err := m.store.Get(id)
			if err == nil {
				db.Status = "stopped"
				db.ProcessID = 0
				m.store.Update(db)
			}
		}
	}

	// Give processes time to shutdown gracefully
	time.Sleep(2 * time.Second)

	// Force kill any remaining processes
	for _, cmd := range m.processes {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
	}

	return nil
}
