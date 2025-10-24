package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yourname/turbo-db/internal/api"
	"github.com/yourname/turbo-db/internal/config"
	"github.com/yourname/turbo-db/internal/database"
	"github.com/yourname/turbo-db/internal/manager"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Turbo DB Server on %s:%s", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Database directory: %s", cfg.DatabasesDir)
	log.Printf("Port range: %d-%d", cfg.PortRangeStart, cfg.PortRangeEnd)

	// Initialize metadata store
	store, err := database.NewStore(cfg.MetaDBPath)
	if err != nil {
		log.Fatalf("Failed to initialize metadata store: %v", err)
	}
	defer store.Close()

	log.Println("Metadata store initialized")

	// Initialize database manager
	mgr, err := manager.NewManager(cfg, store)
	if err != nil {
		log.Fatalf("Failed to initialize database manager: %v", err)
	}

	log.Println("Database manager initialized")

	// Setup API router
	handler := api.NewHandler(mgr)
	router := api.SetupRouter(handler, cfg)

	// Create HTTP server
	addr := fmt.Sprintf("%s:%s", cfg.ServerHost, cfg.ServerPort)
	srv := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Shutdown database manager
	if err := mgr.Shutdown(); err != nil {
		log.Printf("Failed to shutdown database manager: %v", err)
	}

	log.Println("Server exited")
}
